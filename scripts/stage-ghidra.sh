#!/bin/sh
set -eu

version=12.1
release_date=20260513
archive_name=ghidra_${version}_PUBLIC_${release_date}.zip
archive_sha256=aa5cbcbbf48f41ca185fce900e19592f1ade4cd5994eb6e0ede468dac8a6f302
base_url=${AKILIX_GHIDRA_MIRROR:-https://github.com/NationalSecurityAgency/ghidra/releases/download/Ghidra_${version}_build}
cache_dir=${AKILIX_DOWNLOAD_CACHE:-build/downloads}
archive=${cache_dir}/${archive_name}
stage_root=image/kiwi-iso/root
target=${stage_root}/opt/ghidra

mkdir -p "${cache_dir}"
if [ ! -f "${archive}" ]; then
    curl -fsSL -o "${archive}.part" "${base_url}/${archive_name}"
    mv "${archive}.part" "${archive}"
fi
printf '%s  %s\n' "${archive_sha256}" "${archive}" | sha256sum -c -

extract_dir=$(mktemp -d "${cache_dir}/ghidra-${version}.XXXXXX")
trap 'rm -rf -- "${extract_dir}"' EXIT
unzip -q "${archive}" -d "${extract_dir}"
source_dir=${extract_dir}/ghidra_${version}_PUBLIC
test -f "${source_dir}/ghidraRun"
test -f "${source_dir}/LICENSE"

rm -rf -- "${target}"
mkdir -p "${target}"
cp -a "${source_dir}/." "${target}/"
chmod 0755 "${target}/ghidraRun"
install -Dm0755 image/kiwi-iso/overlay/usr/bin/ghidra "${stage_root}/usr/bin/ghidra"
install -Dm0644 image/kiwi-iso/overlay/usr/share/applications/ghidra.desktop "${stage_root}/usr/share/applications/ghidra.desktop"
