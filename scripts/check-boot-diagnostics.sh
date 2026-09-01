#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_root=$(mktemp -d /tmp/akilix-boot-check.XXXXXX)
trap 'rm -rf -- "${test_root}"' EXIT

mkdir -p "${test_root}/usr/share/akilix/boot" \
    "${test_root}/usr/lib/modules/6.12.0-test-default" \
    "${test_root}/etc" "${test_root}/media/boot/grub2" "${test_root}/media/EFI/BOOT"
printf 'akilix-memtest-fixture\n' >"${test_root}/usr/share/akilix/boot/mt86plus-x86_64"
printf 'AKILIX_BUILD_ID=20260901T120000Z\nAKILIX_GIT_COMMIT=abcdef0\nAKILIX_IMAGE_VERSION=0.0.1-m0\n' >"${test_root}/etc/akilix-build"
printf 'menuentry "Akilix Live" { true; }\nmenuentry "Failsafe -- Akilix Live" { true; }\n' >"${test_root}/media/boot/grub2/grub.cfg"
printf 'menuentry "Akilix Live" { true; }\nmenuentry "Failsafe -- Akilix Live" { true; }\n' >"${test_root}/media/EFI/BOOT/grub.cfg"

(cd "${test_root}" && "${repo_dir}/image/kiwi-iso/editbootconfig.sh" "iso:${test_root}/media" 1)

cmp "${test_root}/usr/share/akilix/boot/mt86plus-x86_64" "${test_root}/media/boot/akilix/mt86plus.bin"
cmp "${test_root}/usr/share/akilix/boot/mt86plus-x86_64" "${test_root}/media/boot/akilix/mt86plus.efi"

for config in "${test_root}/media/boot/grub2/grub.cfg" "${test_root}/media/EFI/BOOT/grub.cfg"; do
    for marker in "20260901T120000Z-abcdef0 · 6.12.0-test-default" "Secure Boot must be off" "UEFI Firmware Settings" "Memory test (Memtest86+)" "Restart computer" "Power off computer"; do
        grep -Fq "${marker}" "${config}"
    done
    if command -v grub2-script-check >/dev/null 2>&1; then
        grub2-script-check "${config}"
    elif command -v grub-script-check >/dev/null 2>&1; then
        grub-script-check "${config}"
    fi
done

mkdir -p "${test_root}/empty-media"
if (cd "${test_root}" && "${repo_dir}/image/kiwi-iso/editbootconfig.sh" "iso:${test_root}/empty-media" 1) >/dev/null 2>&1; then
    echo "boot diagnostics hook accepted media without grub.cfg" >&2
    exit 1
fi

echo "boot diagnostics check: BIOS and EFI configurations valid"
