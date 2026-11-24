package io.github.touchbridge.mobile

import android.app.Activity
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.pm.ActivityInfo
import android.content.pm.ApplicationInfo
import android.content.pm.PackageInfo
import android.content.pm.PackageInstaller
import android.content.pm.PackageManager
import android.content.pm.Signature
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.os.Build
import android.os.VibrationAttributes
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager
import android.provider.Settings
import android.util.Base64
import android.util.Log
import android.view.WindowManager
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
import java.io.File
import java.io.IOException
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.HttpURLConnection
import java.net.Inet4Address
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.NetworkInterface
import java.net.SocketTimeoutException
import java.net.URL
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import java.security.MessageDigest
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
    @Volatile private var availableUpdate: AvailableUpdate? = null
    private val updateCheckRunning = AtomicBoolean(false)
    private val updateDownloading = AtomicBoolean(false)
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
        private const val REQUEST_INSTALL_PERMISSION = 4103
        private const val UPDATE_API_URL = "https://api.github.com/repos/Spont-427/touchbridge-controller/releases/latest"
        private const val UPDATE_INSTALL_ACTION = "io.github.touchbridge.mobile.UPDATE_INSTALL_STATUS"
        private const val UPDATE_MAX_BYTES = 200L * 1024 * 1024
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
                    // Some older file providers do not classify .json files as
                    // application/json, which greys them out. A wildcard with
                    // preferred MIME types keeps them selectable; the web UI
                    // still validates the chosen file before importing it.
                    type = "*/*"
                    putExtra(Intent.EXTRA_MIME_TYPES, arrayOf(
                        "application/json",
                        "text/json",
                        "text/plain",
                        "application/octet-stream"
                    ))
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
        fun setKeepScreenOn(enabled: Boolean) {
            mainHandler.post {
                if (enabled) {
                    window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
                } else {
                    window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
                }
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
        fun checkForUpdates(): Boolean {
            if (updateDownloading.get()) {
                postUpdateStatus("error", "An update download is already in progress.", availableUpdate?.version)
                return false
            }
            if (!updateCheckRunning.compareAndSet(false, true)) return false
            mainHandler.post { postUpdateStatus("checking", "Checking GitHub Releases…") }
            worker.execute {
                try {
                    val latest = fetchLatestUpdate()
                    if (isNewerVersion(latest.version, appVersion())) {
                        availableUpdate = latest
                        postUpdateStatus("available", "Version ${latest.version} is available.", latest.version)
                    } else {
                        availableUpdate = null
                        postUpdateStatus("current", "Version ${appVersion()} is the latest stable release.", appVersion())
                    }
                } catch (error: Exception) {
                    Log.w("TouchBridgeUpdate", "Update check failed", error)
                    availableUpdate = null
                    postUpdateStatus("error", updateErrorMessage(error), null)
                } finally {
                    updateCheckRunning.set(false)
                }
            }
            return true
        }

        @JavascriptInterface
        fun installUpdate(): Boolean {
            val update = availableUpdate ?: run {
                postUpdateStatus("error", "Check for updates again before installing.", null)
                return false
            }
            if (updateDownloading.get()) return true
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && !packageManager.canRequestPackageInstalls()) {
                postUpdateStatus("permission", "Allow TouchBridge to install unknown apps, then tap Download and install again.", update.version)
                mainHandler.post {
                    try {
                        startActivityForResult(
                            Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES, Uri.parse("package:$packageName")),
                            REQUEST_INSTALL_PERMISSION
                        )
                    } catch (error: Exception) {
                        postUpdateStatus("error", "The install-permission screen could not be opened.", update.version)
                    }
                }
                return true
            }
            if (!updateDownloading.compareAndSet(false, true)) return true
            postUpdateStatus("downloading", "Starting download…", update.version, 0)
            worker.execute {
                try {
                    val apk = downloadUpdate(update)
                    verifyDownloadedUpdate(apk, update)
                    postUpdateStatus("installing", "Opening the Android installer…", update.version)
                    installDownloadedApk(apk)
                } catch (error: Exception) {
                    Log.w("TouchBridgeUpdate", "Update install failed", error)
                    postUpdateStatus("error", updateErrorMessage(error), update.version)
                } finally {
                    updateDownloading.set(false)
                }
            }
            return true
        }

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

    private fun appVersionCode(): Long = try {
        val info = packageManager.getPackageInfo(packageName, 0)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) info.longVersionCode else {
            @Suppress("DEPRECATION")
            info.versionCode.toLong()
        }
    } catch (_: Exception) {
        0L
    }

    private fun packageVersionCode(info: PackageInfo): Long {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) info.longVersionCode else {
            @Suppress("DEPRECATION")
            info.versionCode.toLong()
        }
    }

    private fun fetchLatestUpdate(): AvailableUpdate {
        val release = readGitHubRelease(UPDATE_API_URL)
        if (release.optBoolean("draft") || release.optBoolean("prerelease")) {
            throw IOException("The latest release is not a stable release.")
        }
        val version = stableVersion(release.optString("tag_name"))
            ?: throw IOException("The latest release has no stable version tag.")
        val debugBuild = (applicationInfo.flags and ApplicationInfo.FLAG_DEBUGGABLE) != 0
        val expectedName = if (debugBuild) "TouchBridge-$version-debug.apk" else "TouchBridge-$version.apk"
        val assets = release.optJSONArray("assets") ?: JSONArray()
        var downloadUrl = ""
        for (index in 0 until assets.length()) {
            val asset = assets.optJSONObject(index) ?: continue
            if (asset.optString("name") != expectedName) continue
            downloadUrl = asset.optString("browser_download_url")
            break
        }
        if (downloadUrl.isBlank()) {
            throw IOException("The latest stable release does not include a compatible APK yet.")
        }
        if (!isTrustedReleaseAssetUrl(downloadUrl, expectedName)) {
            throw IOException("The release APK download link is not trusted.")
        }
        return AvailableUpdate(version, downloadUrl, expectedName)
    }

    private fun readGitHubRelease(url: String): JSONObject {
        val connection = URL(url).openConnection() as HttpURLConnection
        try {
            connection.requestMethod = "GET"
            connection.connectTimeout = 10_000
            connection.readTimeout = 15_000
            connection.instanceFollowRedirects = true
            connection.setRequestProperty("Accept", "application/vnd.github+json")
            connection.setRequestProperty("User-Agent", "TouchBridgeAndroid/${appVersion()}")
            val code = connection.responseCode
            if (code != HttpURLConnection.HTTP_OK) {
                val detail = connection.errorStream?.bufferedReader(StandardCharsets.UTF_8)?.use { it.readText().take(500) }
                Log.w("TouchBridgeUpdate", "GitHub API HTTP $code: ${detail.orEmpty()}")
                throw IOException(if (code == 403) "GitHub rate limit reached. Try again later." else "GitHub returned HTTP $code.")
            }
            if (connection.url.protocol != "https") throw IOException("The update check did not use HTTPS.")
            val body = connection.inputStream.bufferedReader(StandardCharsets.UTF_8).use { it.readText() }
            if (body.length > 2_000_000) throw IOException("The release information was too large.")
            return JSONObject(body)
        } finally {
            connection.disconnect()
        }
    }

    private fun stableVersion(tag: String): String? {
        val match = Regex("^v?(\\d+)\\.(\\d+)\\.(\\d+)$").matchEntire(tag.trim()) ?: return null
        return "${match.groupValues[1]}.${match.groupValues[2]}.${match.groupValues[3]}"
    }

    private fun isNewerVersion(candidate: String, current: String): Boolean {
        val next = versionParts(candidate)
        val installed = versionParts(current)
        for (index in 0 until maxOf(next.size, installed.size)) {
            val left = next.getOrElse(index) { 0 }
            val right = installed.getOrElse(index) { 0 }
            if (left != right) return left > right
        }
        return false
    }

    private fun versionParts(version: String): List<Int> {
        return version.trim().removePrefix("v").split('.').map { part ->
            part.takeWhile(Char::isDigit).toIntOrNull() ?: 0
        }
    }

    private fun isTrustedReleaseAssetUrl(url: String, expectedName: String): Boolean {
        return try {
            val uri = Uri.parse(url)
            uri.scheme == "https" &&
                uri.host == "github.com" &&
                uri.path?.startsWith("/Spont-427/touchbridge-controller/releases/download/") == true &&
                uri.lastPathSegment == expectedName
        } catch (_: Exception) {
            false
        }
    }

    private fun downloadUpdate(update: AvailableUpdate): File {
        val directory = File(cacheDir, "updates")
        if (!directory.exists() && !directory.mkdirs()) throw IOException("The update cache could not be created.")
        directory.listFiles()?.forEach { it.delete() }
        val outputFile = File(directory, update.assetName)
        val connection = URL(update.downloadUrl).openConnection() as HttpURLConnection
        try {
            connection.requestMethod = "GET"
            connection.connectTimeout = 15_000
            connection.readTimeout = 30_000
            connection.instanceFollowRedirects = true
            connection.setRequestProperty("Accept", "application/octet-stream")
            connection.setRequestProperty("User-Agent", "TouchBridgeAndroid/${appVersion()}")
            val code = connection.responseCode
            if (code != HttpURLConnection.HTTP_OK) throw IOException("The update download returned HTTP $code.")
            if (connection.url.protocol != "https") throw IOException("The update download did not use HTTPS.")
            val total = connection.contentLength.toLong()
            if (total > UPDATE_MAX_BYTES) throw IOException("The update file is too large.")
            var downloaded = 0L
            var lastProgress = -5
            connection.inputStream.use { input ->
                outputFile.outputStream().use { output ->
                    val buffer = ByteArray(64 * 1024)
                    while (true) {
                        val count = input.read(buffer)
                        if (count < 0) break
                        downloaded += count
                        if (downloaded > UPDATE_MAX_BYTES) throw IOException("The update file is too large.")
                        output.write(buffer, 0, count)
                        if (total > 0) {
                            val progress = ((downloaded * 100) / total).toInt()
                            if (progress >= lastProgress + 5) {
                                lastProgress = progress
                                postUpdateStatus("downloading", "Downloading update…", update.version, progress)
                            }
                        }
                    }
                }
            }
            if (downloaded == 0L) throw IOException("The update download was empty.")
            postUpdateStatus("downloading", "Download complete. Verifying…", update.version, 100)
            return outputFile
        } catch (error: Exception) {
            outputFile.delete()
            throw error
        } finally {
            connection.disconnect()
        }
    }

    private fun verifyDownloadedUpdate(file: File, update: AvailableUpdate) {
        val flag = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            PackageManager.GET_SIGNING_CERTIFICATES
        } else {
            @Suppress("DEPRECATION")
            PackageManager.GET_SIGNATURES
        }
        val archive = packageManager.getPackageArchiveInfo(file.absolutePath, flag)
            ?: throw IOException("The downloaded file is not a valid APK.")
        if (archive.packageName != packageName) throw IOException("The update is for a different app.")
        val archivedVersion = archive.versionName?.removePrefix("v").orEmpty()
        if (archivedVersion.isNotBlank() && archivedVersion != update.version) {
            throw IOException("The downloaded APK version does not match the release.")
        }
        if (packageVersionCode(archive) <= appVersionCode()) {
            throw IOException("The downloaded update is not newer than the installed app.")
        }
        val current = packageManager.getPackageInfo(packageName, flag)
        if (signatureFingerprints(archive) != signatureFingerprints(current)) {
            throw IOException("The update is not signed by the same developer key. Install it manually from the release page.")
        }
    }

    private fun signatureFingerprints(info: PackageInfo): Set<String> {
        val signatures: Array<Signature> = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            info.signingInfo?.apkContentsSigners ?: info.signingInfo?.signingCertificateHistory ?: emptyArray()
        } else {
            @Suppress("DEPRECATION")
            info.signatures ?: emptyArray()
        }
        return signatures.map { signature ->
            MessageDigest.getInstance("SHA-256").digest(signature.toByteArray())
                .joinToString("") { byte -> "%02x".format(byte.toInt() and 0xff) }
        }.toSet()
    }

    private fun installDownloadedApk(file: File) {
        val installer = packageManager.packageInstaller
        val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL)
        params.setSize(file.length())
        val sessionId = installer.createSession(params)
        try {
            installer.openSession(sessionId).use { session ->
                file.inputStream().use { input ->
                    session.openWrite("base.apk", 0, file.length()).use { output ->
                        input.copyTo(output)
                        session.fsync(output)
                    }
                }
                val callback = Intent(this, UpdateInstallReceiver::class.java).setAction(UPDATE_INSTALL_ACTION)
                // PackageInstaller fills the result extras into this intent, so
                // the PendingIntent must remain mutable on Android 12+.
                val pendingFlags = PendingIntent.FLAG_UPDATE_CURRENT or
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) PendingIntent.FLAG_MUTABLE else 0
                val pending = PendingIntent.getBroadcast(this, sessionId, callback, pendingFlags)
                session.commit(pending.intentSender)
            }
        } catch (error: Exception) {
            installer.abandonSession(sessionId)
            throw error
        }
    }

    private fun postUpdateStatus(status: String, message: String, version: String? = availableUpdate?.version, progress: Int? = null) {
        val payload = JSONObject().apply {
            put("status", status)
            put("message", message)
            if (!version.isNullOrBlank()) put("version", version)
            if (progress != null) put("progress", progress)
        }
        val script = "window.TouchBridge && window.TouchBridge.onUpdateStatus($payload);"
        mainHandler.post { webView?.evaluateJavascript(script, null) }
    }

    private fun updateErrorMessage(error: Exception): String {
        val message = error.message.orEmpty()
        return when {
            message.contains("signed by the same", ignoreCase = true) -> message
            message.contains("rate limit", ignoreCase = true) -> message
            message.contains("does not include", ignoreCase = true) -> message
            message.contains("not newer", ignoreCase = true) -> message
            message.contains("different app", ignoreCase = true) -> message
            message.contains("version does not match", ignoreCase = true) -> message
            message.contains("too large", ignoreCase = true) -> message
            message.contains("not a valid APK", ignoreCase = true) -> message
            else -> "Could not install the update. Check your connection and try again."
        }
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
        if (requestCode == REQUEST_INSTALL_PERMISSION) {
            val granted = Build.VERSION.SDK_INT < Build.VERSION_CODES.O || packageManager.canRequestPackageInstalls()
            postUpdateStatus(
                if (granted) "available" else "permission",
                if (granted) "Install permission granted. Tap Download and install to continue." else "TouchBridge still needs permission to install unknown apps.",
                availableUpdate?.version
            )
            return
        }
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
        window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        pendingFileChooser?.onReceiveValue(null)
        pendingFileChooser = null
        pendingExportContent = null
        scanRunning.set(false)
        updateCheckRunning.set(false)
        updateDownloading.set(false)
        worker.shutdownNow()
        webView?.apply { stopLoading(); removeJavascriptInterface("TouchBridgeAndroid"); destroy() }
        webView = null
        super.onDestroy()
    }
}

private data class AvailableUpdate(
    val version: String,
    val downloadUrl: String,
    val assetName: String
)

class UpdateInstallReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        when (intent.getIntExtra(PackageInstaller.EXTRA_STATUS, PackageInstaller.STATUS_FAILURE)) {
            PackageInstaller.STATUS_PENDING_USER_ACTION -> {
                val confirmation = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                    intent.getParcelableExtra(Intent.EXTRA_INTENT, Intent::class.java)
                } else {
                    @Suppress("DEPRECATION")
                    intent.getParcelableExtra(Intent.EXTRA_INTENT)
                }
                if (confirmation == null) {
                    Toast.makeText(context, "The Android installer could not be opened.", Toast.LENGTH_LONG).show()
                    return
                }
                confirmation.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                try {
                    context.startActivity(confirmation)
                } catch (_: Exception) {
                    Toast.makeText(context, "The Android installer could not be opened.", Toast.LENGTH_LONG).show()
                }
            }
            PackageInstaller.STATUS_SUCCESS ->
                Toast.makeText(context, "TouchBridge updated.", Toast.LENGTH_SHORT).show()
            else -> {
                val message = intent.getStringExtra(PackageInstaller.EXTRA_STATUS_MESSAGE)
                    ?.takeIf { it.isNotBlank() }
                    ?: "The Android installer did not complete the update."
                Toast.makeText(context, message, Toast.LENGTH_LONG).show()
            }
        }
    }
}
