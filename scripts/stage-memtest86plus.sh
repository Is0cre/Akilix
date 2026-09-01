#!/bin/sh
set -eu

version=8.10
archive_name=mt86plus_${version}.binaries.zip
archive_sha256=7e6c5162cb84ab959aeb9d13c9cfd6976b0dec3b34936b73820b20c55eb26c29
base_url=${AKILIX_MEMTEST_MIRROR:-https://memtest.org/download/v${version}}
cache_dir=${AKILIX_DOWNLOAD_CACHE:-build/downloads}
archive=${cache_dir}/${archive_name}
target=image/kiwi-iso/root/usr/share/akilix/boot/mt86plus-x86_64

mkdir -p "${cache_dir}"
if [ ! -f "${archive}" ]; then
    curl -fsSL -o "${archive}.part" "${base_url}/${archive_name}"
    mv "${archive}.part" "${archive}"
fi
printf '%s  %s\n' "${archive_sha256}" "${archive}" | sha256sum -c -

extract_dir=$(mktemp -d "${cache_dir}/memtest-${version}.XXXXXX")
trap 'rm -rf -- "${extract_dir}"' EXIT
unzip -q "${archive}" mt86p_810_x86_64 -d "${extract_dir}"
install -Dm0644 "${extract_dir}/mt86p_810_x86_64" "${target}"
