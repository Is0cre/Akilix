# Akilix Branding

Branding is a reproducible image input. Canonical assets belong in the source
tree and are copied into images through the KIWI overlay; generated files must
not be edited inside `build/`.

Akilix branding must not obscure the openSUSE Leap foundation or imply that
an early development image is ready for production forensic work.

## Boot and session surfaces

The current development live image has these branding surfaces:

- the `Akilix Live` image and boot-menu display name;
- `/etc/issue.d/50-akilix.issue` before console login;
- the `greetd`/`tuigreet` login and Sway session;
- a static Plymouth boot screen;
- CLI version and help output.

Plymouth is activated by `quiet splash`, but pressing Esc exposes boot
diagnostics. The theme has no animation, font dependency, or external resource.
Its script renderer and Dracut integration come from the approved openSUSE OBS
`Base:System` source; exact RPM identities are recorded in
`repositories/boot-plymouth-lock.json`.

## Canonical artwork inputs

Approved artwork is stored under `branding/`. The canonical and OS-facing
inputs are:

- `source/akilix-master.png` — canonical full raster artwork;
- `source/akilix-mark-master.png` — canonical compact raster artwork;
- `source/akilix-wordmark.svg` — editable wordmark;
- `os/grub/background-1920x1080.png` — 16:9 bootloader background;
- `os/grub/logo.png` — compact bootloader mark;
- `os/plymouth/logo.png` — centered boot-splash mark;
- `os/wallpaper/akilix-3840x2160.png` — Sway wallpaper.

SVG files must not contain scripts, remote references, embedded fonts, or data
URIs. Relative references must remain inside `branding/`. Keep source artwork
and its license/provenance record beside the assets. Do not derive branding
from trademarks that Akilix is not authorized to redistribute.

## Integration and testing

1. Validate the canonical asset set with `make branding-check`.
2. Stage generated image-overlay copies with `make branding-stage`.
3. Test GRUB and Plymouth in BIOS and UEFI modes, including the Esc diagnostic
   path and a low-resolution display.
4. Confirm `greetd` appears after Plymouth exits and logout returns to it.

The boot menu must remain usable if graphical assets fail to load. Branding
must not suppress boot diagnostics, evidence-safety warnings, or the fact that
the image is an early development build.
