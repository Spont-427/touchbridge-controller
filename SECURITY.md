# Security notes

TouchBridge is an early-alpha LAN input-control application. Do not expose it to the public internet.

- Default HTTP/WebSocket traffic is unencrypted. The six-digit PIN is visible to anyone able to sniff local traffic. Use only trusted private networks unless TLS is configured with a browser-trusted certificate.
- Android encrypts trusted-server bearer tokens at rest with an Android Keystore AES-GCM key; the server stores only token hashes. Default HTTP/WebSocket transport remains unencrypted, so tokens can still be observed in transit on an untrusted LAN.
- PINs expire after ten minutes and failed attempts are rate-limited. This reduces casual guessing but does not turn a six-digit code into strong cryptographic transport security.
- Discovery on UDP 8766 is unauthenticated and returns the responding computer name and port. It does not grant control.
- A connected client can inject keyboard and mouse events into the active desktop session. Keep the server stopped when not in use.
- Linux input access is restricted to `/dev/uinput`; the server should run as the logged-in user, not as root.
- Input injection behaviour under UAC/elevated Windows applications, locked desktops, Wayland compositors, multi-seat systems and sandboxed desktops varies.

Report suspected vulnerabilities privately to the project maintainers before public disclosure.
