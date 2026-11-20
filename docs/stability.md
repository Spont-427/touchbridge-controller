# Stability and automated quality checks

TouchBridge uses automated checks and review requirements for input-injection and security-sensitive code.

## Checks in CI

Run `./scripts/check.sh` locally for the same deterministic test, static-analysis and short-fuzz gates. The script does not rewrite source files.

- `gofmt` format verification.
- `go test -race ./...` for Go unit and integration tests.
- `go vet ./...`.
- Short fuzz runs against the JSON protocol decoder and WebSocket frame parser.
- JavaScript syntax and pure controller-math tests under Node.js.
- Installer/build shell syntax checks.
- Linux amd64/arm64 and Windows amd64 cross-compilation.
- Android debug APK build using JDK 17, Android SDK 35 and Gradle 8.9.
- CodeQL static analysis for Go and JavaScript/TypeScript.
- `govulncheck` from the Go project for reachable known Go vulnerabilities.
- Dependabot pull requests for Go modules, GitHub Actions and Gradle build tooling. The Gradle action uses its basic open-source cache provider.

Every proposed change must be reviewed and pass CI before merge. Security-sensitive changes require particular care. The project does not use automatic merging or unattended source modification. The local build environment had no outbound DNS access, so the `govulncheck` binary could not be downloaded or run locally; CI runs it in a network-enabled runner.

## Input safety invariants

- Only one authenticated controller may hold the connection slot.
- Failed PIN attempts are rate-limited.
- All remote input messages are size- and range-checked before injection.
- Gamepad axes/triggers are finite and bounded; button names are allow-listed and duplicates rejected.
- Abandoned connections and page focus loss release held state.
- Application shutdown closes upgraded WebSocket connections and waits for handlers to release input before destroying virtual devices.
- Profile changes, layout import/reset and entering editor mode clear any held gamepad state before rebuilding the control DOM.
- WebSocket congestion retains only bounded movement and the latest gamepad snapshot instead of replaying stale input indefinitely.
- Local process status is loopback-only because it includes the current PIN.
- Virtual-device creation is optional and must fail closed without disabling mouse/keyboard control.

## What tests cannot establish

Cross-compilation and unit tests do not prove that real desktop input works on every kernel, compositor, game, Windows security configuration or Android manufacturer. Linux `uinput`, Windows `SendInput`, Wi-Fi latency, sleep/resume and real gamepad enumeration must be tested on physical devices before a stable release.
