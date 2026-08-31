# Akilix Desktop Direction

## Decision

Akilix intends to provide a lightweight, keyboard-first Sway desktop with
tiling enabled by default. The desktop is an operator UX layer; it must not
duplicate workbook, scope, evidence, or provenance business logic from the Go
backend.

The plain Linux console remains a supported recovery and diagnostic path.
Failure of the compositor, graphical greeter, bar, or launcher must not make
the CLI or workbook data inaccessible.

## Leap 16 package constraint

As of August 2026, Sway and Waybar builds for Leap 16 are published through the
openSUSE Build Service `X11:Wayland` project, not the official Leap 16 release
repository:

- <https://software.opensuse.org/package/sway?locale=en>
- <https://software.opensuse.org/download/package?package=waybar&project=X11%3AWayland>

The repository is an explicitly approved experimental input for the development
ISO. Akilix records its URL and signing-key fingerprint and keeps the resolved
x86_64 RPM identities in `repositories/desktop-sway-lock.json`. A local mirror
or Akilix OBS project remains preferred for reproducible release builds.

The repository identity is recorded as an approved development-image source in
`repositories/repositories.json`. Inspect it locally with:

    akilix repository show x11-wayland-leap-16

This inspection performs no network access. Enabling the source in the KIWI
description does not add it to an already-running host.

## Minimal session

The initial graphical session should contain only enough software to provide a
comfortable workstation:

- Sway compositor and tiling window manager;
- Xwayland for applications that still require X11;
- Foot or another small Wayland-native terminal;
- Sway's separately packaged native `swaybar` with a local Akilix workbook
  status stream;
- Fuzzel or an equivalently small application launcher;
- FeatherPad as a lightweight graphical text editor (`Super+Shift+N`);
- PCManFM-Qt as a lightweight graphical file manager (`Super+Shift+F`), with
  removable-media automount and autorun disabled;
- `nnn` as the terminal file navigator (`Super+Shift+O`);
- `btop` as the on-demand system and network monitor (`Super+Shift+B`);
- BlueZ and Blueman for operator-controlled Bluetooth (`Super+Shift+P`), backed
  by PipeWire and WirePlumber for audio devices;
- Mako for local notifications;
- wl-clipboard;
- Grim and Slurp for explicit operator-requested screenshots;
- a minimal PolicyKit agent where graphical elevation workflows require it;
- the Akilix wallpaper and icon assets already stored under `branding/`.
- an offline-by-default Weathr ASCII animation for the idle screen;

The development ISO runs `greetd` with the terminal-native `tuigreet` greeter
on tty1. Authentication explicitly starts Sway; automatic account login is not
enabled. Logging out or a failed session returns to the greeter. SSH sessions
and tty2+ remain text-only recovery paths. The greeter does not remember the
last username or selected session. Its centered Akilix terminal artwork is read
from `/etc/issue`; the graphite/green/amber terminal palette, zoned clock, Sway
session selector, and explicit power menu are provided by pinned `tuigreet`
0.9.1 options. F2 opens sessions, F10 opens power actions, and F12 opens the
authenticated command selector.

No component may silently start scans, listeners, packet capture, terminal
recording, telemetry, cloud synchronization, or workbook uploads.

BlueZ is available through D-Bus, but Akilix does not automatically power
Bluetooth adapters, begin discovery, become discoverable, or start pairing.
Discoverable and pairable modes have 180-second limits when the operator
enables them. The Blueman tray applet is not autostarted.

The image includes `iw`, regulatory data, tcpdump, wavemon, horst, and
Wireshark for explicit wireless observation. Akilix does not automatically put
an interface into monitor mode, change its channel or regulatory domain, or
start packet capture. Those actions affect host networking and require an
operator-selected workflow and any necessary explicit elevation. Monitor-mode
support depends on the physical adapter, driver, and firmware.

Aircrack-ng is supplied by the audited openSUSE `security/16.0` OBS repository.
Its presence does not authorize deauthentication, credential attacks, or any
other active operation. Hydra is available for explicitly authorized managed
invocations but has no automatic action in the workbook TUI. Kismet is not
published for x86_64 by that repository at the pinned revision and is therefore
not claimed or requested by the image.

Bluetooth diagnostics also include the deprecated BlueZ compatibility tools
for older adapters and protocols plus the supplementary BlueZ firmware set.
They are operator-invoked; no test utility or background scanner is enabled.

The reverse-engineering baseline includes binutils, Radare2, Binwalk, YARA,
Sleuth Kit, EWF tools, TestDisk, and ExifTool. Ghidra 12.1 is staged from the
official release archive with its published SHA-256 digest and runs on Leap's
OpenJDK 21. It is available as `ghidra` and through Fuzzel, but is never started
automatically.

## Default interaction model

The primary modifier should be the Super key. Initial bindings should include:

- `Super+Enter` — terminal;
- `Ctrl+Alt+T` — terminal fallback for VM consoles that intercept Super;
- `Super+D` — launcher;
- `Alt+F2` — launcher fallback;
- `Super+Shift+N` — FeatherPad;
- `Super+Shift+F` — PCManFM-Qt;
- `Super+Shift+O` — `nnn` in Foot;
- `Super+Shift+B` — `btop` in Foot;
- `Super+Shift+P` — Bluetooth manager;
- `Super+Shift+Q` — close focused window;
- `Super+1` through `Super+9` — select workspace;
- `Super+Shift+1` through `Super+Shift+9` — move a window;
- `Super+H/J/K/L` and arrow keys — move focus;
- `Super+Shift+H/J/K/L` — move a tile;
- `Super+F` — toggle fullscreen;
- `Super+Shift+Space` — toggle floating;
- `Super+Shift+E` — show an explicit logout prompt;
- `Print` — operator-requested screenshot only.

The live session opens one Foot terminal at startup. This gives an otherwise
empty tiling workspace an obvious first action without depending on the VM
console forwarding the host's Super key. It does not run a Akilix command,
select a workbook, or initiate network activity.

Sway displays the staged Akilix 3840×2160 wallpaper on every output using
`fill` scaling. Foot uses the matching graphite palette at 78% opacity so the
branding is clearly visible behind terminal work while preserving text
contrast. Per-user Foot or Sway configuration may override these system
defaults.

## Idle screen and locking

After five idle minutes, Sway launches Weathr fullscreen in simulated rainy
night mode. Simulation is deliberate: the idle screen makes no weather,
geocoding, or IP-location request. Input dismisses that animation. After ten
idle minutes, and before suspend, `swaylock` provides the actual authentication
boundary; the animation must never be described as a lock screen. `Super+Shift+R`
starts the animation manually and `Super+Shift+L` locks immediately.

The development ISO stages the upstream GPL-3.0-or-later Weathr 1.4.0 x86_64
binary using its published SHA-256 digest and installs its license. The download
cache and both artifact origins are overrideable for local mirrors. No unpinned
installer or `latest` URL is executed during the image build.

The image explicitly installs Noto Sans for desktop text, Noto Sans Mono for
Foot, Noto Sans Symbols 2 as a technical-symbol fallback, and Noto Color Emoji
for the command strip and terminal UI. Font selection must not depend on
whatever sparse fallback happens to enter the dependency closure. Exact font
RPM identities are recorded in
`repositories/desktop-fonts-lock.json`.

Opening the desktop or bar must not open a workbook or perform network
activity. Workbook widgets may display local state only after the operator has
explicitly selected a workbook.

The native command strip runs `/usr/bin/akilix bar stream` using the i3bar JSON
protocol. The absolute path makes startup independent of an interactive shell's
`PATH`; each status frame is flushed immediately. Entering a workbook TUI
atomically marks that workbook active beneath
`$XDG_RUNTIME_DIR/akilix/`. The strip derives declared scope and running or
failed managed invocations from canonical local files. It does not claim packet
drops, hashing queues, or other metrics Akilix does not yet record. No daemon,
shared global UID, network request, or telemetry service is involved.

## Go terminal experience

The `akilix` Go backend remains the source of truth. Terminal comfort should
grow through:

- fast shell completion without network lookups;
- timestamped, durable Zsh history with an explicit secret-safe omission path;
- consistent human and JSON output;
- discoverable help and safe command previews;
- readable status, scope, logging, and provenance summaries;
- a Bubble Tea workbook workspace that calls the same packages as the CLI.

The TUI must not become a second implementation of workbook logic and must not
introduce hidden monitoring or automatic engagement actions. The interactive
action model returns typed selections; the parent workbook loop owns all scope,
playbook, logging, and lifecycle operations.

When attached to a terminal, the workbook TUI clears and redraws the dashboard
after each completed action so its panels remain at stable screen positions.
Redirected output and test writers do not receive terminal-control sequences.
Actions are shown as an aligned vertical command palette with highlighted key
badges and semantic colors. Arrow keys and `j`/`k` wrap through the list;
Enter selects the focused action, and direct action hotkeys remain available.
Non-terminal input retains a deterministic line-oriented path for automation
and tests. Bubble Tea v2 establishes Go 1.25 as the source-build floor; the ISO
contains the compiled Akilix binary rather than a Go runtime.

## Image acceptance tests

Before the desktop becomes the ISO default, test at minimum:

- BIOS and UEFI boot retain a usable console fallback;
- Sway starts unprivileged for the live operator;
- software rendering works in a VM;
- representative Intel, AMD, and NVIDIA behavior is documented;
- keyboard layout, HiDPI, multi-monitor, suspend, and resume behavior;
- terminal, launcher, bar, notifications, clipboard, and screenshots;
- no unexpected listeners or Akilix network activity after login;
- opening and validating a workbook remains passive;
- original evidence paths are never exposed as writable convenience mounts;
- logout returns to a controlled login or console state.

The first graphical image remains a development image until these checks are
recorded.
