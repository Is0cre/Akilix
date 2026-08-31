#!/bin/sh
set -eu

version=1.4.0
binary_sha256=7b9761f2b97183e4ac546bbcda45719d63a05a95aed84d996580b6975ba26213
license_sha256=3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986
base_url=${PENSUSE_WEATHR_MIRROR:-https://github.com/Veirt/weathr/releases/download/v${version}}
source_url=${PENSUSE_WEATHR_SOURCE_MIRROR:-https://raw.githubusercontent.com/Veirt/weathr/v${version}}
cache_dir=${PENSUSE_DOWNLOAD_CACHE:-build/downloads}
binary=${cache_dir}/weathr-${version}-linux-amd64
license=${cache_dir}/weathr-${version}-LICENSE

mkdir -p "${cache_dir}"
if [ ! -f "${binary}" ]; then
    curl -fsSL -o "${binary}.part" "${base_url}/weathr-linux-amd64"
    mv "${binary}.part" "${binary}"
fi
printf '%s  %s\n' "${binary_sha256}" "${binary}" | sha256sum -c -

if [ ! -f "${license}" ]; then
    curl -fsSL -o "${license}.part" "${source_url}/LICENSE"
    mv "${license}.part" "${license}"
fi
printf '%s  %s\n' "${license_sha256}" "${license}" | sha256sum -c -

install -Dm0755 "${binary}" image/kiwi-iso/root/usr/bin/weathr
install -Dm0644 "${license}" image/kiwi-iso/root/usr/share/licenses/weathr/LICENSE
