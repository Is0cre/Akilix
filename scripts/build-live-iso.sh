#!/bin/sh
set -eu

if [ "$(id -u)" -eq 0 ]; then
    need_sudo=0
else
    need_sudo=1
fi

if [ -z "${AKILIX_LIVE_PASSWORD+x}" ]; then
    printf '%s' 'Live akilix password: ' >&2
    old_stty=$(stty -g)
    trap 'stty "$old_stty" 2>/dev/null || true; printf "\n" >&2' EXIT HUP INT TERM
    stty -echo
    IFS= read -r AKILIX_LIVE_PASSWORD
    stty "$old_stty"
    trap - EXIT HUP INT TERM
    printf '\n' >&2
    export AKILIX_LIVE_PASSWORD
fi

if [ "$need_sudo" -eq 1 ]; then
    : "${AKILIX_BUILD_ID:=$(date -u +%Y%m%dT%H%M%SZ)}"
    : "${AKILIX_GIT_COMMIT:=$(git rev-parse --short=12 HEAD)}"
    export AKILIX_BUILD_ID AKILIX_GIT_COMMIT
    exec sudo --preserve-env=AKILIX_LIVE_PASSWORD,AKILIX_BUILD_ID,AKILIX_GIT_COMMIT make kiwi-iso
fi
exec make kiwi-iso
