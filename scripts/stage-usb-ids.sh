#!/bin/sh
set -eu

version=2026.06.26
database_sha256=f5a48b0cc8dae1607c2f0bae6b8dc13f2ecef69dbeaeaf34b4be8e280d34dba4
source_url=${AKILIX_USB_IDS_MIRROR:-https://usb-ids.gowdy.us/usb.ids}
cache_dir=${AKILIX_DOWNLOAD_CACHE:-build/downloads}
database=${cache_dir}/usb.ids-${version}

mkdir -p "${cache_dir}"
if [ ! -f "${database}" ]; then
    curl -fsSL -o "${database}.part" "${source_url}"
    mv "${database}.part" "${database}"
fi
printf '%s  %s\n' "${database_sha256}" "${database}" | sha256sum -c -
grep -Fqx '# Version: 2026.06.26' "${database}"

# usbutils already owns this standard path. The pinned upstream snapshot is
# staged as data only; no lookup is performed over the network at runtime.
install -Dm0644 "${database}" image/kiwi-iso/root/usr/share/hwdata/usb.ids
install -Dm0644 docs/licenses/USB_IDS.md image/kiwi-iso/root/usr/share/licenses/akilix-usb-ids/NOTICE.md
