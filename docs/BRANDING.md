# PenSUSE Branding

Branding is a reproducible image input. Canonical assets belong in the source
tree and are copied into images through the KIWI overlay; generated files must
not be edited inside `build/`.

PenSUSE branding must not obscure the openSUSE Leap foundation or imply that
an early development image is ready for production forensic work.

## Initial console image

The current M0 live image is console-only. Its initial branding surface is:

- the `PenSUSE Live` image and boot-menu display name;
- `/etc/issue.d/50-pensuse.issue` before console login;
- the existing `PenSUSE Operator` account name;
- CLI version and help output.

No desktop, display manager, or Plymouth package is installed. Adding artwork
must not silently add those subsystems or introduce new services.

## Canonical artwork inputs

Approved artwork is stored under `branding/`. The canonical and OS-facing
inputs are:

- `source/pensuse-master.png` — canonical full raster artwork;
- `source/pensuse-mark-master.png` — canonical compact raster artwork;
- `source/pensuse-wordmark.svg` — editable wordmark;
- `os/grub/background-1920x1080.png` — 16:9 bootloader background;
- `os/grub/logo.png` — compact bootloader mark;
- `os/wallpaper/pensuse-3840x2160.png` — future desktop wallpaper.

SVG files must not contain scripts, remote references, embedded fonts, or data
URIs. Relative references must remain inside `branding/`. Keep source artwork
and its license/provenance record beside the assets. Do not derive branding
from trademarks that PenSUSE is not authorized to redistribute.

## Planned integration order

1. Validate and record the approved logo provenance and license.
2. Install shared artwork under `/usr/share/pensuse/branding/`.
3. Test the KIWI/GRUB theme in BIOS, UEFI, serial-console, and text-only paths.
4. Add Plymouth only if boot testing justifies the extra package and initrd
   behavior.
5. Integrate wallpaper and desktop identity only after a desktop environment is
   selected explicitly.

The boot menu must remain usable if graphical assets fail to load. Branding
must not suppress boot diagnostics, evidence-safety warnings, or the fact that
the image is an early development build.
