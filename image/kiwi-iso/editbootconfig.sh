#!/bin/sh
set -eu

filesystem=${1:?boot filesystem argument required}
media=${filesystem#iso:}
source_binary=usr/share/akilix/boot/mt86plus-x86_64
test -f "${source_binary}"
test -d "${media}"

install -Dm0644 "${source_binary}" "${media}/boot/akilix/mt86plus.bin"
install -Dm0644 "${source_binary}" "${media}/boot/akilix/mt86plus.efi"

find "${media}" -type f -name grub.cfg -print | while IFS= read -r config; do
    cat >>"${config}" <<'EOF'

# Akilix offline boot diagnostics. Memtest86+ is upstream-pinned but unsigned.
if [ "$grub_platform" = "efi" ]; then
    menuentry "Memory test (Memtest86+; Secure Boot must be off)" --class recovery --unrestricted {
        insmod chain
        chainloader ($root)/boot/akilix/mt86plus.efi
    }
    menuentry "UEFI Firmware Settings" --class recovery --unrestricted { fwsetup; }
else
    menuentry "Memory test (Memtest86+)" --class recovery --unrestricted {
        linux16 ($root)/boot/akilix/mt86plus.bin
    }
fi
menuentry "Restart computer" --class recovery --unrestricted { reboot; }
menuentry "Power off computer" --class recovery --unrestricted { halt; }
EOF
done

find "${media}" -type f -name grub.cfg -print -quit | grep -q . || {
    echo "Akilix boot customization: no grub.cfg found" >&2
    exit 1
}
