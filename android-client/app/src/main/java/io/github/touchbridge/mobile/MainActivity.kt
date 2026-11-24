package io.github.touchbridge.mobile

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.content.pm.ActivityInfo
import android.content.pm.ApplicationInfo
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.os.Build
import android.os.VibrationAttributes
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager
import android.util.Base64
import android.webkit.JavascriptInterface
import android.webkit.MimeTypeMap
import android.webkit.WebChromeClient
import android.webkit.WebResourceRequest
import android.webkit.WebResourceResponse
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import android.webkit.ValueCallback
import android.widget.Toast
import android.net.Uri
import org.json.JSONArray
import org.json.JSONObject
import java.io.ByteArrayInputStream
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.Inet4Address
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.NetworkInterface
import java.net.SocketTimeoutException
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import java.util.Collections
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicBoolean
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties

/**
 * Native Android host for the self-contained TouchBridge controller UI.
 * The app UI is bundled locally; only controller input travels to a PC server.
 */
class MainActivity : Activity() {
    private val worker = Executors.newSingleThreadExecutor()
    private val mainHandler = Handler(Looper.getMainLooper())
    private val scanRunning = AtomicBoolean(false)
    private var webView: WebView? = null
    private var immersiveLandscape = false
    private var lightTheme = false
    private var pendingFileChooser: ValueCallback<Array<Uri>>? = null
    private var pendingExportContent: String? = null
    private var pendingExportFilename = "touchbridge-layout.json"
    private val preferences by lazy {
        getSharedPreferences("touchbridge.devices.v1", Context.MODE_PRIVATE)
    }
    private val storeLock = Any()

    companion object {
        private const val APP_HOST = "touchbridge.appassets.local"
        private const val APP_ORIGIN = "https://$APP_HOST"
        private const val TRUST_KEY_ALIAS = "touchbridge.device.tokens.v1"
        private const val TRUST_PREF_KEY = "trusted_servers"
        private const val DISCOVERY_PORT = 8766
        private const val DISCOVERY_PROBE = "TOUCHBRIDGE_DISCOVER_V1"
        private const val REQUEST_IMPORT_LAYOUT = 4101
        private const val REQUEST_EXPORT_LAYOUT = 4102
        private val THEME_LIGHT_BG = 0xFFF4F4F1.toInt()
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        applySystemBars()
        createControllerWebView()
    }

    @Suppress("SetJavaScriptEnabled")
    private fun createControllerWebView() {
        if ((applicationInfo.flags and ApplicationInfo.FLAG_DEBUGGABLE) != 0) {
            // Debug builds only: lets the bundled page be inspected over CDP.
            WebView.setWebContentsDebuggingEnabled(true)
        }
        val view = WebView(this)
        view.setBackgroundColor(android.graphics.Color.BLACK)
        view.isVerticalScrollBarEnabled = false
        view.isHorizontalScrollBarEnabled = false
        view.overScrollMode = WebView.OVER_SCROLL_NEVER
        view.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            databaseEnabled = false
            allowFileAccess = false
            allowContentAccess = false
            javaScriptCanOpenWindowsAutomatically = false
            setSupportMultipleWindows(false)
            mixedContentMode = WebSettings.MIXED_CONTENT_ALWAYS_ALLOW
            cacheMode = WebSettings.LOAD_DEFAULT
            userAgentString = "$userAgentString TouchBridgeAndroid/${appVersion()}"
        }
        view.addJavascriptInterface(AndroidBridge(), "TouchBridgeAndroid")
        view.webChromeClient = object : WebChromeClient() {
            override fun onShowFileChooser(
                webView: WebView?,
                filePathCallback: ValueCallback<Array<Uri>>?,
                fileChooserParams: WebChromeClient.FileChooserParams?
            ): Boolean {
                pendingFileChooser?.onReceiveValue(null)
                pendingFileChooser = filePathCallback
                val intent = Intent(Intent.ACTION_OPEN_DOCUMENT).apply {
                    addCategory(Intent.CATEGORY_OPENABLE)
                    type = "application/json"
                    addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                }
                return try {
                    startActivityForResult(intent, REQUEST_IMPORT_LAYOUT)
                    true
                } catch (_: Exception) {
                    pendingFileChooser?.onReceiveValue(null)
                    pendingFileChooser = null
                    Toast.makeText(this@MainActivity, "No file picker is available.", Toast.LENGTH_SHORT).show()
                    true
                }
            }
        }
        view.webViewClient = object : WebViewClient() {
            override fun shouldInterceptRequest(view: WebView?, request: WebResourceRequest): WebResourceResponse? {
                val uri = request.url
                if (!isAppAssetUri(uri)) return null
                return localAssetResponse(uri.path ?: "/index.html")
            }

            override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest): Boolean {
                // Keep the WebView on its bundled origin. No remote page gets access
                // to the Android bridge object.
                return !isAppAssetUri(request.url)
            }

            @Suppress("DEPRECATION")
            override fun shouldOverrideUrlLoading(view: WebView?, url: String): Boolean {
                return try {
                    !isAppAssetUri(android.net.Uri.parse(url))
                } catch (_: Exception) {
                    true
                }
            }
        }
        webView = view
        setContentView(view)
        view.loadUrl("$APP_ORIGIN/index.html")
    }

    private fun localAssetResponse(rawPath: String): WebResourceResponse {
        val path = rawPath.removePrefix("/").ifBlank { "index.html" }
        val safe = path.split('/').all { it.isNotBlank() && it != "." && it != ".." } && !path.contains('\\')
        if (!safe) return response(404, "Not Found", "text/plain", "Not found")
        return try {
            val stream = assets.open("controller/$path")
            val extension = path.substringAfterLast('.', "").lowercase()
            val mime = when (extension) {
                "html" -> "text/html"
                "css" -> "text/css"
                "js" -> "application/javascript"
                "svg" -> "image/svg+xml"
                "json", "webmanifest" -> "application/json"
                "png" -> "image/png"
                "jpg", "jpeg" -> "image/jpeg"
                "woff2" -> "font/woff2"
                else -> MimeTypeMap.getSingleton().getMimeTypeFromExtension(extension) ?: "application/octet-stream"
            }
            WebResourceResponse(mime, "UTF-8", 200, "OK", mapOf("Cache-Control" to "no-cache", "X-Content-Type-Options" to "nosniff"), stream)
        } catch (_: Exception) {
            response(404, "Not Found", "text/plain", "Asset not found")
        }
    }

    private fun response(code: Int, reason: String, mime: String, body: String): WebResourceResponse {
        return WebResourceResponse(
            mime,
            "UTF-8",
            code,
            reason,
            mapOf("Cache-Control" to "no-store", "X-Content-Type-Options" to "nosniff"),
            ByteArrayInputStream(body.toByteArray(StandardCharsets.UTF_8))
        )
    }

    private inner class AndroidBridge {
        @JavascriptInterface
        fun getSavedServersJson(): String = synchronized(storeLock) {
            val stored = try { JSONArray(preferences.getString(TRUST_PREF_KEY, "[]") ?: "[]") } catch (_: Exception) { JSONArray() }
            val output = JSONArray()
            val items = (0 until stored.length())
                .mapNotNull { stored.optJSONObject(it) }
                .sortedByDescending { it.optLong("lastUsed", 0L) }
            for (item in items) {
                val url = item.optString("url")
                if (!isValidServerUrl(url)) continue
                val tokenCipher = item.optString("tokenCipher")
                val decrypted = if (tokenCipher.isNotBlank()) decryptToken(tokenCipher) else ""
                val token = decrypted.takeIf { it.matches(Regex("^[a-fA-F0-9]{64}$")) } ?: ""
                output.put(JSONObject().apply {
                    put("url", url)
                    put("name", item.optString("name", "TouchBridge computer"))
                    put("serverId", item.optString("serverId"))
                    put("token", token)
                    put("lastUsed", item.optLong("lastUsed", 0L))
                })
            }
            output.toString()
        }

        @JavascriptInterface
        fun saveTrustedServer(url: String, name: String, serverId: String, token: String): Boolean {
            if (!isValidServerUrl(url)) return false
            if (token.isNotBlank() && !token.matches(Regex("^[a-fA-F0-9]{64}$"))) return false
            if (serverId.isNotBlank() && !serverId.matches(Regex("^[a-fA-F0-9]{32}$"))) return false
            val normalisedName = name.trim().take(80).ifBlank { "TouchBridge computer" }
            val encrypted = if (token.isNotBlank()) encryptToken(token) ?: return false else ""
            synchronized(storeLock) {
                val old = try { JSONArray(preferences.getString(TRUST_PREF_KEY, "[]") ?: "[]") } catch (_: Exception) { JSONArray() }
                val next = JSONArray()
                for (i in 0 until old.length()) {
                    val item = old.optJSONObject(i) ?: continue
                    val sameId = serverId.isNotBlank() && item.optString("serverId") == serverId
                    val sameUrl = item.optString("url") == url
                    if (!sameId && !sameUrl) next.put(item)
                }
                next.put(JSONObject().apply {
                    put("url", url)
                    put("name", normalisedName)
                    put("serverId", serverId)
                    put("tokenCipher", encrypted)
                    put("lastUsed", System.currentTimeMillis())
                })
                while (next.length() > 16) next.remove(0)
                return preferences.edit().putString(TRUST_PREF_KEY, next.toString()).commit()
            }
        }

        @JavascriptInterface
        fun forgetTrustedServer(serverIdOrUrl: String): Boolean {
            synchronized(storeLock) {
                val old = try { JSONArray(preferences.getString(TRUST_PREF_KEY, "[]") ?: "[]") } catch (_: Exception) { JSONArray() }
                val next = JSONArray()
                for (i in 0 until old.length()) {
                    val item = old.optJSONObject(i) ?: continue
                    if (item.optString("serverId") != serverIdOrUrl && item.optString("url") != serverIdOrUrl) next.put(item)
                }
                return preferences.edit().putString(TRUST_PREF_KEY, next.toString()).commit()
            }
        }

        @JavascriptInterface
        fun scanServers() {
            if (!scanRunning.compareAndSet(false, true)) return
            worker.execute {
                val found = discoverServers()
                val payload = JSONArray()
                found.forEach { payload.put(it) }
                val quoted = payload.toString()
                mainHandler.post {
                    webView?.evaluateJavascript(
                        "window.TouchBridge && window.TouchBridge.onServersFound($quoted);",
                        null
                    )
                    scanRunning.set(false)
                }
            }
        }

        @JavascriptInterface
        fun setScreenMode(mode: String) {
            mainHandler.post {
                if (mode == "landscape") enterImmersiveLandscape() else leaveImmersiveLandscape()
            }
        }

        @JavascriptInterface
        fun setThemeMode(mode: String) {
            mainHandler.post {
                lightTheme = mode.equals("light", ignoreCase = true)
                webView?.setBackgroundColor(if (lightTheme) THEME_LIGHT_BG else android.graphics.Color.BLACK)
                applySystemBars()
            }
        }

        @JavascriptInterface
        fun getAppVersion(): String = appVersion()

        @JavascriptInterface
        fun openRepository() {
            mainHandler.post {
                try {
                    startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("https://github.com/Spont-427/touchbridge-controller")))
                } catch (_: Exception) {
                    Toast.makeText(this@MainActivity, "No browser is available to open the repository.", Toast.LENGTH_SHORT).show()
                }
            }
        }

        @JavascriptInterface
        fun performHapticFeedback() {
            try {
                // Prefer the platform's tuned click waveform. A 9 ms raw one-shot
                // is below the perceptible threshold on many phone motors, which
                // made the press feel like it only reacted after a long-press.
                val vibrator = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                    (getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as? VibratorManager)?.defaultVibrator
                } else {
                    @Suppress("DEPRECATION")
                    getSystemService(Context.VIBRATOR_SERVICE) as? Vibrator
                } ?: return
                if (!vibrator.hasVibrator()) return
                when {
                    Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q -> {
                        val effect = VibrationEffect.createPredefined(VibrationEffect.EFFECT_CLICK)
                        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                            vibrator.vibrate(effect, VibrationAttributes.createForUsage(VibrationAttributes.USAGE_TOUCH))
                        } else {
                            vibrator.vibrate(effect)
                        }
                    }
                    Build.VERSION.SDK_INT >= Build.VERSION_CODES.O ->
                        vibrator.vibrate(VibrationEffect.createOneShot(22L, 200))
                    else -> {
                        @Suppress("DEPRECATION")
                        vibrator.vibrate(22L)
                    }
                }
            } catch (_: Exception) {
                // Haptics are optional; a device without a vibrator must still work.
            }
        }

        @JavascriptInterface
        fun saveTextFile(filename: String, contents: String): Boolean {
            if (contents.toByteArray(StandardCharsets.UTF_8).size > 64 * 1024) return false
            val safeName = filename.replace(Regex("[^A-Za-z0-9._-]"), "_").take(80).ifBlank { "touchbridge-layout.json" }
            mainHandler.post {
                pendingExportContent = contents
                pendingExportFilename = safeName
                val intent = Intent(Intent.ACTION_CREATE_DOCUMENT).apply {
                    addCategory(Intent.CATEGORY_OPENABLE)
                    type = "application/json"
                    putExtra(Intent.EXTRA_TITLE, pendingExportFilename)
                    addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION or Intent.FLAG_GRANT_READ_URI_PERMISSION)
                }
                try {
                    startActivityForResult(intent, REQUEST_EXPORT_LAYOUT)
                } catch (_: Exception) {
                    pendingExportContent = null
                    Toast.makeText(this@MainActivity, "No save location is available.", Toast.LENGTH_SHORT).show()
                }
            }
            return true
        }
    }

    private fun appVersion(): String = try {
        @Suppress("DEPRECATION")
        packageManager.getPackageInfo(packageName, 0).versionName ?: "0.0.0"
    } catch (_: Exception) {
        "0.0.0"
    }

    private fun isAppAssetUri(uri: android.net.Uri): Boolean {
        return uri.scheme.equals("https", ignoreCase = true) &&
            uri.host.equals(APP_HOST, ignoreCase = true) &&
            uri.port == -1 && uri.userInfo == null
    }

    private fun isValidServerUrl(value: String): Boolean {
        return try {
            val uri = android.net.Uri.parse(value)
            (uri.scheme == "http" || uri.scheme == "https") && !uri.host.isNullOrBlank() && uri.userInfo == null
        } catch (_: Exception) {
            false
        }
    }

    private fun discoverServers(): List<JSONObject> {
        val found = linkedMapOf<String, JSONObject>()
        val socket = try {
            DatagramSocket(null).apply {
                reuseAddress = true
                broadcast = true
                soTimeout = 180
                bind(InetSocketAddress(0))
            }
        } catch (_: Exception) {
            return emptyList()
        }

        val broadcasts = linkedSetOf<InetAddress>()
        try { broadcasts.add(InetAddress.getByName("255.255.255.255")) } catch (_: Exception) { }
        try {
            Collections.list(NetworkInterface.getNetworkInterfaces())
                .filter { it.isUp && !it.isLoopback }
                .forEach { network ->
                    network.interfaceAddresses.forEach { iface ->
                        if (iface.address is Inet4Address && iface.broadcast != null) broadcasts.add(iface.broadcast)
                    }
                }
        } catch (_: Exception) { /* Broadcast fallback is still attempted. */ }

        try {
            val probe = DISCOVERY_PROBE.toByteArray(StandardCharsets.UTF_8)
            broadcasts.forEach { address ->
                try { socket.send(DatagramPacket(probe, probe.size, address, DISCOVERY_PORT)) } catch (_: Exception) { }
            }
            val deadline = System.currentTimeMillis() + 1900L
            val buffer = ByteArray(512)
            while (System.currentTimeMillis() < deadline) {
                val packet = DatagramPacket(buffer, buffer.size)
                try {
                    socket.receive(packet)
                } catch (_: SocketTimeoutException) {
                    continue
                }
                val body = String(packet.data, packet.offset, packet.length, StandardCharsets.UTF_8)
                val lines = body.split('\n')
                if (lines.size < 3 || lines[0] != "TOUCHBRIDGE_SERVER_V1") continue
                val port = lines[1].toIntOrNull() ?: continue
                if (port !in 1..65535) continue
                val ip = packet.address.hostAddress ?: continue
                val scheme = lines.getOrNull(4)?.trim()?.takeIf { it == "https" } ?: "http"
                val serverUrl = "$scheme://$ip:$port"
                val serverId = lines.getOrNull(3)?.trim().orEmpty().take(128)
                found[if (serverId.isNotBlank()) serverId else serverUrl] = JSONObject().apply {
                    put("name", lines[2].trim().take(80).ifBlank { "TouchBridge computer" })
                    put("address", ip)
                    put("port", port)
                    put("url", serverUrl)
                    put("serverId", serverId)
                }
            }
        } catch (_: Exception) {
            // Keep results received before a transient interface/network error.
        } finally {
            socket.close()
        }
        return found.values.toList()
    }

    private fun getOrCreateTrustKey(): SecretKey {
        val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        (keyStore.getKey(TRUST_KEY_ALIAS, null) as? SecretKey)?.let { return it }
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore")
        generator.init(
            KeyGenParameterSpec.Builder(
                TRUST_KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setRandomizedEncryptionRequired(true)
                .build()
        )
        return generator.generateKey()
    }

    private fun encryptToken(token: String): String? = try {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, getOrCreateTrustKey())
        val encrypted = cipher.doFinal(token.toByteArray(StandardCharsets.UTF_8))
        val combined = cipher.iv + encrypted
        Base64.encodeToString(combined, Base64.NO_WRAP)
    } catch (_: Exception) {
        null
    }

    private fun decryptToken(value: String): String {
        return try {
            val combined = Base64.decode(value, Base64.NO_WRAP)
            if (combined.size <= 12) "" else {
                val iv = combined.copyOfRange(0, 12)
                val encrypted = combined.copyOfRange(12, combined.size)
                val cipher = Cipher.getInstance("AES/GCM/NoPadding")
                cipher.init(Cipher.DECRYPT_MODE, getOrCreateTrustKey(), GCMParameterSpec(128, iv))
                String(cipher.doFinal(encrypted), StandardCharsets.UTF_8)
            }
        } catch (_: Exception) {
            ""
        }
    }

    @Suppress("DEPRECATION")
    private fun enterImmersiveLandscape() {
        immersiveLandscape = true
        requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
        applyImmersiveFlags()
    }

    @Suppress("DEPRECATION")
    private fun applyImmersiveFlags() {
        window.decorView.systemUiVisibility = (
            android.view.View.SYSTEM_UI_FLAG_FULLSCREEN or
                android.view.View.SYSTEM_UI_FLAG_HIDE_NAVIGATION or
                android.view.View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY or
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_STABLE or
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN or
                android.view.View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
            )
    }

    @Suppress("DEPRECATION")
    private fun leaveImmersiveLandscape() {
        immersiveLandscape = false
        requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        window.decorView.systemUiVisibility = android.view.View.SYSTEM_UI_FLAG_LAYOUT_STABLE
        applySystemBars()
    }

    @Suppress("DEPRECATION")
    private fun applySystemBars() {
        // Android 15+ draws edge-to-edge for targetSdk 35, so the colours are
        // advisory; the icon appearance flags are what keep the bars readable.
        window.statusBarColor = if (lightTheme) THEME_LIGHT_BG else android.graphics.Color.BLACK
        window.navigationBarColor = if (lightTheme) THEME_LIGHT_BG else android.graphics.Color.BLACK
        var flags = window.decorView.systemUiVisibility
        flags = if (lightTheme) {
            var next = flags or android.view.View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                next = next or android.view.View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR
            }
            next
        } else {
            var next = flags and android.view.View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR.inv()
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                next = next and android.view.View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR.inv()
            }
            next
        }
        window.decorView.systemUiVisibility = flags
    }

    @Suppress("DEPRECATION")
    override fun onWindowFocusChanged(hasFocus: Boolean) {
        super.onWindowFocusChanged(hasFocus)
        if (hasFocus && immersiveLandscape) applyImmersiveFlags()
    }

    @Suppress("DEPRECATION")
    override fun onBackPressed() {
        val view = webView
        if (view == null) {
            super.onBackPressed()
            return
        }
        view.evaluateJavascript("String(Boolean(window.TouchBridge && window.TouchBridge.androidBack()))") { result ->
            if (result == "false" || result == "null") super.onBackPressed()
        }
    }

    @Deprecated("Deprecated in Android API; retained for compatibility with the minSdk 23 file picker flow")
    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        if (requestCode == REQUEST_IMPORT_LAYOUT) {
            val callback = pendingFileChooser
            pendingFileChooser = null
            val uri = if (resultCode == RESULT_OK) data?.data else null
            callback?.onReceiveValue(if (uri == null) null else arrayOf(uri))
            return
        }
        if (requestCode == REQUEST_EXPORT_LAYOUT) {
            val contents = pendingExportContent
            pendingExportContent = null
            val uri = if (resultCode == RESULT_OK) data?.data else null
            if (uri == null || contents == null) {
                Toast.makeText(this, "Export cancelled.", Toast.LENGTH_SHORT).show()
                return
            }
            try {
                contentResolver.openOutputStream(uri, "wt")?.use { stream ->
                    stream.write(contents.toByteArray(StandardCharsets.UTF_8))
                } ?: throw IllegalStateException("Unable to open the selected file")
                Toast.makeText(this, "Layout exported.", Toast.LENGTH_SHORT).show()
            } catch (error: Exception) {
                Toast.makeText(this, "Could not save layout: ${error.message}", Toast.LENGTH_LONG).show()
            }
            return
        }
        super.onActivityResult(requestCode, resultCode, data)
    }

    override fun onDestroy() {
        pendingFileChooser?.onReceiveValue(null)
        pendingFileChooser = null
        pendingExportContent = null
        scanRunning.set(false)
        worker.shutdownNow()
        webView?.apply { stopLoading(); removeJavascriptInterface("TouchBridgeAndroid"); destroy() }
        webView = null
        super.onDestroy()
    }
}
