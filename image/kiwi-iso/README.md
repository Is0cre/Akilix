# Akilix M0 ISO definition

This KIWI definition produces an openSUSE Leap 16 ISO suitable for a first
USB/Ventoy boot test. It intentionally contains platform prerequisites only;
it is not a security-tool distribution.

Build on a privileged Leap 16 build host:

```sh
make kiwi-iso-prompt
```

The helper asks for the password without echoing it. The password is required
at build time, converted to a SHA-512 crypt hash, and never written to Git.
The resulting image provides the regular `akilix`
operator account; root remains unavailable for ordinary console login.

The target builds the current CLI and shell completions into the image overlay
before KIWI runs. The result is written under `build/kiwi-iso/`. If that path
already exists, the target first preserves it as
`build/kiwi-iso.previous-YYYYMMDDTHHMMSSZ`; KIWI never reuses an existing image
root. Verify the generated checksum before copying the ISO to Ventoy media. A Ventoy boot test
must verify that the system reaches the expected target, does not start Akilix
listeners, and preserves the documented passive-opening behavior.

Every successful `make kiwi-iso` now runs
`scripts/check-iso-boot-payload.sh` against the finished ISO and writes an
adjacent `.iso.sha256` file. The check extracts the kernel and initrd, requires
matching kernel/module releases, verifies `/init` resolves to systemd, and
verifies systemd plus its dynamic loader are x86-64 ELF payloads. Before a USB
test, run `sha256sum -c akilix-m0-iso.x86_64-0.0.1.iso.sha256` on the copied
files. A kernel release shown on the target that differs from the verified ISO
proves that a stale or different boot payload was selected.

The build stages `/etc/akilix-build` with a UTC build ID, Git commit, and image
version. GRUB shows the build ID, commit, and kernel release in both normal and
failsafe labels, and the interactive MOTD repeats the identity after login.
After verification, Make creates a storage-efficient hard link named like
`akilix-0.0.1-m0-20260901T120000Z-abcdef012345.x86_64.iso`, writes its portable
checksum sidecar, and records the path in `build/kiwi-iso/LATEST-ISO`. Copy the
uniquely named ISO and its sidecar to Ventoy; do not rename it back to the
ambiguous KIWI filename.

The image display name, console banner, GRUB menu, Plymouth screen, and Sway
session are branded through tracked KIWI inputs. Artwork requirements are
documented in `docs/BRANDING.md`.

`make kiwi-iso` validates the canonical branding kit and stages the Akilix
GRUB background, theme, shared logo, and asset license into the KIWI overlay.
The graphical boot menu retains GRUB's text-console fallback and must be tested
in both BIOS and UEFI modes before the image is published.

The boot menu also stages the official Memtest86+ 8.10 unified x86_64 binary,
checksum-pinned by `scripts/stage-memtest86plus.sh`, and adds memory test,
firmware setup (UEFI), restart, and power-off entries through KIWI's supported
`editbootconfig` hook. Memtest86+ is not Secure-Boot signed; its UEFI menu label
states that Secure Boot must be disabled. Normal Akilix boot remains signed and
does not disable or bypass firmware policy.
`make boot-diagnostics-check` exercises the KIWI hook against separate BIOS and
EFI GRUB configurations, validates copied payload identity and menu markers,
uses GRUB's native syntax checker when available, and verifies that a missing
GRUB configuration fails the build hook.

The static Plymouth theme uses the approved `Base:System` OBS source because
Leap 16 OSS does not currently carry the required package set. Its repository
key and exact RPM identities are recorded in the repository manifests. Press
Esc during boot to expose diagnostics; the theme does not replace that path.

The development build includes the minimal Sway session and `greetd`/`tuigreet`
from the approved `X11:Wayland` OBS source. Account auto-login is disabled.
The greeter owns tty1 and starts Sway only after authentication; logout or
session failure returns to the greeter, while tty2+ remain recovery consoles.
The login screen uses the compact Akilix terminal mark, UTC clock, themed
prompt, F2 session menu, and F10 power menu without remembering operator names.
The session opens one Foot terminal as its visible starting point;
`Ctrl+Alt+T` remains available when a VM or host compositor intercepts the
Super key. Exact OBS RPM identities are recorded in
`repositories/desktop-sway-lock.json`.
The split `swaybar` RPM is installed explicitly; installing `sway` alone does
not provide `/usr/bin/swaybar` on this Leap repository.

The approved openSUSE `security/16.0` OBS source supplies Aircrack-ng and Hydra.
Its signing key, repository revision, primary metadata digest, and exact RPM
identities are recorded in `repositories/security-tools-lock.json`. Kismet is
not present for x86_64 at that revision and is intentionally omitted instead of
leaving an unresolvable KIWI package request.

The live image is useful before any OCI profile is imported: `iputils`
provides `ping`, native Nmap provides operator-requested discovery and service
inspection, and tcpdump/ethtool/iproute2 provide local diagnostics.
NetworkManager is the selected desktop network stack, with built-in DHCP for
automatic wired connectivity and `nmtui` for keyboard-first connection setup.
Wicked is explicitly disabled to prevent two network managers competing for an
interface. DHCP is ordinary link configuration; Akilix still performs no
automatic scanning, listener creation, telemetry, or workbook network action.
The Nmap RPM identity and the `network:utilities/16.0` signing-key identity are
recorded in the repository manifests.

`make network-image-check` statically verifies the KIWI package selection for
NetworkManager, `nmtui`, iputils, and Nmap, and confirms that the effective
systemd presets enable NetworkManager while disabling Wicked. The check reads
only repository-owned image inputs; it does not inspect or change host network
state.

Ghidra is the documented exception to RPM-only application delivery because no
Leap 16 package repository currently publishes it. `scripts/stage-ghidra.sh`
downloads the official 12.1 archive into the reusable build cache and verifies
SHA-256 `aa5cbcbbf48f41ca185fce900e19592f1ade4cd5994eb6e0ede468dac8a6f302`
before staging `/opt/ghidra`. Set `AKILIX_GHIDRA_MIRROR` to an internal mirror
or pre-populate `AKILIX_DOWNLOAD_CACHE` for an offline build.

The operator account uses Zsh with the documented Akilix system baseline,
durable extended history, and generated CLI completion. See `docs/SHELL.md`.

After login, validate the container substrate without pulling an image:

```sh
akilix container doctor
```

A passing result confirms that Podman is operating rootless and that the live
operator's user namespace is usable. It does not claim that a tool image has
already been installed.
