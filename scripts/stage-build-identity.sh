#!/bin/sh
set -eu

if [ "$#" -ne 4 ]; then
	echo "usage: $0 BUILD_ID GIT_COMMIT IMAGE_VERSION OUTPUT" >&2
	exit 2
fi
build_id=$1
commit=$2
version=$3
output=$4

case $build_id in *[!0-9TZ]*) echo "invalid build ID" >&2; exit 1;; esac
case $commit in *[!0-9a-f]*) echo "invalid Git commit" >&2; exit 1;; esac
case $version in *[!0-9A-Za-z._-]*) echo "invalid image version" >&2; exit 1;; esac
[ ${#commit} -ge 7 ] || { echo "Git commit identity is too short" >&2; exit 1; }

mkdir -p "$(dirname "$output")"
temporary=${output}.tmp
trap 'rm -f -- "$temporary"' EXIT
{
	printf 'AKILIX_BUILD_ID=%s\n' "$build_id"
	printf 'AKILIX_GIT_COMMIT=%s\n' "$commit"
	printf 'AKILIX_IMAGE_VERSION=%s\n' "$version"
} >"$temporary"
chmod 0644 "$temporary"
mv "$temporary" "$output"
trap - EXIT
