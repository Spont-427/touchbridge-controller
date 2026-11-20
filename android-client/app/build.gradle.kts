plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "io.github.touchbridge.mobile"
    compileSdk = 35

    defaultConfig {
        applicationId = "io.github.touchbridge.mobile"
        minSdk = 23
        targetSdk = 35
        versionCode = 10
        versionName = "0.5.2-alpha"
    }

    // Keep Java and Kotlin bytecode targets aligned.
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }
}
