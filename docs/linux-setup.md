# Linux setup

## One-command installer (recommended)

Extract the TouchBridge release archive, open a terminal in the extracted folder, and run:

```sh
./Install-TouchBridge-Linux.sh
```

The installer asks for administrator permission once, then:

- installs the server at `/usr/local/bin/touchbridge`;
- configures the `uinput` module to load automatically;
- installs a udev rule granting the active desktop user access to the virtual input device (without running the network server as root);
- adds a **TouchBridge Server** application-menu entry;
- starts the server and displays the six-digit pairing PIN.

The installer attempts to start the server with the new `touchbridge` group immediately, without requiring a logout. On distributions where the active session manager does not grant access until re-login, log out and back in once, then start **TouchBridge Server** from your application menu or run `touchbridge` in a terminal.

Uninstall with:

```sh
./scripts/uninstall-linux.sh
```

## Manual setup

TouchBridge uses the kernel `/dev/uinput` interface. Load it with:

```sh
sudo modprobe uinput
```

For a manual setup, install `packaging/linux/70-touchbridge-uinput.rules` into `/etc/udev/rules.d/`, install `packaging/linux/uinput.conf` into `/etc/modules-load.d/`, create a `touchbridge` group, add your account to that group, reload udev rules, and log out/in. The installer performs these steps for you.

## Desktop notes

- The uinput backend works at the kernel input layer and is intended for X11 and Wayland desktops.
- The server runs as the logged-in user, not as root.
- Avoid granting broad read/write access to physical `/dev/input/event*` devices. TouchBridge only needs `/dev/uinput`.
- A desktop's active-seat policy, locked screen, multi-seat setup, and sandbox restrictions can affect injection.
- Keep the firewall limited to the trusted private network. TCP `8765` carries the controller and UDP `8766` is used only for discovery.
