# TouchBridge LAN protocol (0.6.0)

TouchBridge uses a small JSON protocol over a persistent WebSocket connection. The default TCP listener is port `8765`.

## Discovery

The desktop server listens on UDP port `8766` by default. Android sends the exact UTF-8 datagram:

```text
TOUCHBRIDGE_DISCOVER_V1
```

The server responds to the sender's UDP source address with:

```text
TOUCHBRIDGE_SERVER_V1\n<http-port>\n<computer-name>\n<server-id>\n<scheme>
```

The final fields identify the server across DHCP address changes and advertise `http` or `https`. Older clients may ignore the additional fields; newer clients fall back to `http` when talking to older servers. Discovery is unauthenticated and only provides a local server address. It grants no input-control access. Set `-discovery-port 0` to disable it.

## Pairing

The terminal displays a cryptographically generated six-digit PIN. It is valid for ten minutes and is automatically replaced while the server remains running; the terminal prints each renewed PIN. The client sends it as the first WebSocket text message:

```json
{"type":"auth","pin":"012345"}
```

The server limits failed attempts to five per source IP per minute and forty total per minute. Only one controller can be active at a time. Authentication succeeds with:

```json
{"type":"auth_ok","server_version":"0.6.0"}
```

The `server_version` field identifies the connected desktop build; older clients may ignore it, and clients should tolerate its absence when connecting to older servers. The server then sends a `capabilities` message describing whether the selected desktop input backend exposes a virtual gamepad. Otherwise it sends `auth_error`, `rate_limited` or `controller_busy` and closes the connection. The PIN is a short local pairing code, not a replacement for encrypted transport; use only trusted private networks unless TLS is configured.

## Mouse and keyboard input

After authentication, send one JSON object per WebSocket text frame:

```json
{"type":"mouse_move","dx":12,"dy":-4}
{"type":"mouse_button","button":"left","down":true}
{"type":"scroll","amount":-2}
{"type":"key","key":"SPACE","down":true}
{"type":"release_all"}
```

Supported mouse buttons are `left`, `right`, and `middle`. Supported keys are the portable set defined in `internal/input/input.go`. Movement is relative; each `dx` and `dy` is bounded to ±500 and scroll amount to ±20. Frames larger than 4 KiB are rejected.

## Virtual gamepad input

The controller sends complete state snapshots rather than individual axis deltas. This makes reconnects and state recovery deterministic:

```json
{
  "type":"gamepad_state",
  "left_x":0.65,"left_y":-0.20,
  "right_x":0.0,"right_y":0.0,
  "left_trigger":0.8,"right_trigger":0.0,
  "buttons":["A","LB","DPAD_UP"]
}
```

Stick axes are normalized to `[-1, 1]`; triggers are normalized to `[0, 1]`. `buttons` contains the currently held controls from the supported set `A`, `B`, `X`, `Y`, `LB`, `RB`, `BACK`, `START`, `GUIDE`, `LS`, `RS`, and `DPAD_UP`, `DPAD_DOWN`, `DPAD_LEFT`, `DPAD_RIGHT`. Values outside those ranges, non-finite numbers, unknown buttons and duplicate button names are rejected.

Use `{"type":"gamepad_reset"}` to return all axes, triggers and buttons to neutral. A client disconnect and `release_all` also reset the virtual gamepad. Linux exposes an evdev-compatible virtual gamepad through `/dev/uinput` when the kernel module and permissions allow it. Windows exposes an Xbox 360-compatible controller through ViGEmBus when the driver is installed; without it, `capabilities.gamepad` is `false` and only mouse and keyboard input are accepted.

## Keepalive and latency probe

The client sends `{"type":"keepalive"}` every 15 seconds; the server replies `{"type":"keepalive_ack"}`. If no message arrives for 45 seconds, the server closes the session. The server releases all held inputs when the controller disconnects or the idle timeout expires.

For latency measurement, send `{"type":"latency_ping","timestamp":<client-monotonic-ms>}`. The server echoes the timestamp in `latency_pong`. The client calculates round-trip time locally and reports a rolling median, 95th percentile and jitter estimate. The timestamp is only used for measurement; clocks are not synchronized and this is not a one-way latency measurement.

## Local process status

`GET /__touchbridge/status` returns the server identifier, version, current PIN and expiry only when the request originates from a loopback address (`127.0.0.1` or `::1`). It is used by the desktop launcher to recognize an already-running TouchBridge instance and display its current PIN instead of attempting a duplicate bind. Requests from LAN addresses receive HTTP 403. The response is marked `Cache-Control: no-store` and must never be exposed through a reverse proxy.

## Compatibility

This is a new protocol and is not wire-compatible with the supplied Touch-Server executable. Protocol changes are versioned in this document and should be covered by tests before release.
