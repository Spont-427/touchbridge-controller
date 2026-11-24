plugins {
    id("com.android.application")
}

android {
    namespace = "io.github.touchbridge.mobile"
    compileSdk = 35

    defaultConfig {
        applicationId = "io.github.touchbridge.mobile"
        minSdk = 23
        targetSdk = 35
        versionCode = 14
        versionName = "0.8.0"
    }

    // Keep Java and Kotlin bytecode targets aligned.
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlin {
        compilerOptions {
            jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
        }
    }
}
