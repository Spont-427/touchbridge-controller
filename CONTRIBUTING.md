# Contributing

TouchBridge is an independent clean-room implementation. Contributions should preserve the documented protocol and must not copy proprietary code, game HUD artwork or other assets from third-party applications.

## Before opening a pull request

Run the deterministic checks:

```sh
gofmt -w .
go test -race ./...
go vet ./...
node --check internal/web/web/app.js
node --check internal/web/web/gamepad-core.js
node --test internal/web/web/gamepad-core.test.cjs
bash -n Install-TouchBridge-Linux.sh scripts/*.sh
```

The CI workflow also cross-compiles Linux amd64/arm64 and Windows amd64, builds the Android debug APK, and runs CodeQL. Dependabot proposes dependency and action updates.

## Input and security changes

Changes to input injection, pairing, WebSocket handling, virtual-device setup or shutdown must include regression tests. Review every patch and run the full test suite before merging. Never add a driver installer or elevate the whole server without an explicit, documented user consent path and tests for clean uninstall and crash recovery.

## Platform testing

Cross-compilation is not a substitute for testing. Changes to a platform backend should be exercised on a real Windows/Linux host and, where relevant, with SDL/Steam Input, XInput, Android WebView and Wi-Fi discovery.
