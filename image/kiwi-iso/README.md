# PenSUSE M0 ISO definition

This KIWI definition produces an openSUSE Leap 16 ISO suitable for a first
USB/Ventoy boot test. It intentionally contains platform prerequisites only;
it is not a security-tool distribution.

Build on a privileged Leap 16 build host:

```sh
make kiwi-iso-prompt
```

The helper asks for the password without echoing it. The password is required
at build time, converted to a SHA-512 crypt hash, and never written to Git.
The resulting image provides the regular `pensuse`
operator account; root remains unavailable for ordinary console login.

The target builds the current CLI and shell completions into the image overlay
before KIWI runs. The result is written under `build/kiwi-iso/`. Verify the
generated checksum before copying the ISO to Ventoy media. A Ventoy boot test
must verify that the system reaches the expected target, does not start PenSUSE
listeners, and preserves the documented passive-opening behavior.

The image display name and console banner are branded through tracked KIWI
inputs. Artwork requirements and the planned boot-theme integration are
documented in `docs/BRANDING.md`. The current console-only M0 image does not
install a desktop or Plymouth merely to display branding.

`make kiwi-iso` validates the canonical branding kit and stages the PenSUSE
GRUB background, theme, shared logo, and asset license into the KIWI overlay.
The graphical boot menu retains GRUB's text-console fallback and must be tested
in both BIOS and UEFI modes before the image is published.

The next development build includes the minimal Sway session from the approved
`X11:Wayland` OBS source. Account auto-login is intentionally disabled. Sway
starts after the live operator logs in on tty1; exiting or failing Sway returns
to the console. SSH and secondary-console logins remain text-only. Exact OBS
RPM identities are recorded in `repositories/desktop-sway-lock.json`.
