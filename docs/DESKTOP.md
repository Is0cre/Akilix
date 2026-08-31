# PenSUSE Desktop Direction

## Decision

PenSUSE intends to provide a lightweight, keyboard-first Sway desktop with
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
ISO. PenSUSE records its URL and signing-key fingerprint and keeps the resolved
x86_64 RPM identities in `repositories/desktop-sway-lock.json`. A local mirror
or PenSUSE OBS project remains preferred for reproducible release builds.

The repository identity is recorded as an approved development-image source in
`repositories/repositories.json`. Inspect it locally with:

    pensuse repository show x11-wayland-leap-16

This inspection performs no network access. Enabling the source in the KIWI
description does not add it to an already-running host.

## Minimal session

The initial graphical session should contain only enough software to provide a
comfortable workstation:

- Sway compositor and tiling window manager;
- Xwayland for applications that still require X11;
- Foot or another small Wayland-native terminal;
- Waybar with local system and workbook status only;
- Fuzzel or an equivalently small application launcher;
- Mako for local notifications;
- wl-clipboard;
- Grim and Slurp for explicit operator-requested screenshots;
- a minimal PolicyKit agent where graphical elevation workflows require it;
- the PenSUSE wallpaper and icon assets already stored under `branding/`.

The development ISO runs `greetd` with the terminal-native `tuigreet` greeter
on tty1. Authentication explicitly starts Sway; automatic account login is not
enabled. Logging out or a failed session returns to the greeter. SSH sessions
and tty2+ remain text-only recovery paths. The greeter does not remember the
last username or selected session.

No component may silently start scans, listeners, packet capture, terminal
recording, telemetry, cloud synchronization, or workbook uploads.

## Default interaction model

The primary modifier should be the Super key. Initial bindings should include:

- `Super+Enter` — terminal;
- `Ctrl+Alt+T` — terminal fallback for VM consoles that intercept Super;
- `Super+D` — launcher;
- `Alt+F2` — launcher fallback;
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
console forwarding the host's Super key. It does not run a PenSUSE command,
select a workbook, or initiate network activity.

The image explicitly installs Noto Sans for desktop text, Noto Sans Mono for
Foot, and Noto Sans Symbols 2 as a technical-symbol fallback. Font selection
must not depend on whatever sparse fallback happens to enter the dependency
closure. Exact font RPM identities are recorded in
`repositories/desktop-fonts-lock.json`.

Opening the desktop or bar must not open a workbook or perform network
activity. Workbook widgets may display local state only after the operator has
explicitly selected a workbook.

## Go terminal experience

The `pensuse` Go backend remains the source of truth. Terminal comfort should
grow through:

- fast shell completion without network lookups;
- timestamped, durable Zsh history with an explicit secret-safe omission path;
- consistent human and JSON output;
- discoverable help and safe command previews;
- readable status, scope, logging, and provenance summaries;
- a panel-based workbook workspace today and a future interactive Go TUI that
  calls the same packages as the CLI.

A future TUI must not become a second implementation of workbook logic and
must not introduce hidden monitoring or automatic engagement actions.

## Image acceptance tests

Before the desktop becomes the ISO default, test at minimum:

- BIOS and UEFI boot retain a usable console fallback;
- Sway starts unprivileged for the live operator;
- software rendering works in a VM;
- representative Intel, AMD, and NVIDIA behavior is documented;
- keyboard layout, HiDPI, multi-monitor, suspend, and resume behavior;
- terminal, launcher, bar, notifications, clipboard, and screenshots;
- no unexpected listeners or PenSUSE network activity after login;
- opening and validating a workbook remains passive;
- original evidence paths are never exposed as writable convenience mounts;
- logout returns to a controlled login or console state.

The first graphical image remains a development image until these checks are
recorded.
