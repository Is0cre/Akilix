# PenSUSE branding assets

This package faithfully derives Linux and software branding assets from the supplied PenSUSE concept artwork. The detailed monitor-lizard identity is preserved as raster master artwork rather than replaced by a low-quality geometric vector approximation.

## Canonical artwork

- `source/pensuse-master.png` — full natural-history monitor lizard master.
- `source/pensuse-mark-master.png` — compact raised monitor-head/scope master.
- `source/pensuse-wordmark.svg` — editable PenSUSE wordmark typography.

The concept boards in `reference/` are art direction and are not distributable logos. The monitor lizard is an independent PenSUSE project identity and is not the openSUSE Geeko logo. These assets do not imply endorsement by SUSE or openSUSE.

## Identity

The monitor lizard communicates observation, patience, awareness, deliberate action, and scope discipline. Its long tail forms an incomplete scope circle; three amber nodes indicate provenance and traceability. The incomplete circle intentionally avoids claiming perfect containment.

The palette is Deep Graphite `#0B1114`, Graphite `#151A1D`, Bone `#E8E6DD`, Monitor Green `#657A3E`, Leaf Green `#8EAD55`, Signal Amber `#C68A2B`, and Slate `#4A5358`. Amber is an attention/provenance accent only.

The canonical spelling is `PenSUSE`. The canonical tagline is `Security work with provenance.` Uppercase display text may use `SECURITY WORK WITH PROVENANCE.` Use the full master for large logos, web headers, social cards, boot, installer, wallpapers, documentation, and stickers. Use the compact master for application icons, favicons, and small UI. Do not shrink the full animal to make micro icons.

Maintain at least 2X clear space around logo compositions, where X is the monitor eye diameter. Preserve aspect ratio, do not rotate or stretch the artwork, close the scope circle, move nodes, add glow, use noisy backgrounds, alter capitalization, or add security clichés such as skulls and crosshairs.

## Build and review

From this directory:

```sh
make              # all raster derivatives and terminal assets
make web          # web and favicon derivatives
make os           # Plymouth, GRUB, installer, wallpaper
make icons        # desktop icon derivatives
make terminal     # hand-authored terminal artwork
make validate     # structural, integrity, dimension, alpha, ASCII, ANSI checks
make clean        # remove generated raster files
```

The build uses local Python 3/Pillow and ImageMagick only; it performs no network downloads. Resizing uses Lanczos except for nearest-neighbour favicon review tiles. `preview.png` is a generated contact sheet containing the primary logo, compact mark, enlarged 16px/32px views, social card, Plymouth, GRUB, wallpaper, and terminal review samples. Inspect it before shipping.
