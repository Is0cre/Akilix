#!/bin/sh
set -eu

if [ "$(id -u)" -eq 0 ]; then
    need_sudo=0
else
    need_sudo=1
fi

if [ -z "${PENSUSE_LIVE_PASSWORD+x}" ]; then
    printf '%s' 'Live pensuse password: ' >&2
    old_stty=$(stty -g)
    trap 'stty "$old_stty" 2>/dev/null || true; printf "\n" >&2' EXIT HUP INT TERM
    stty -echo
    IFS= read -r PENSUSE_LIVE_PASSWORD
    stty "$old_stty"
    trap - EXIT HUP INT TERM
    printf '\n' >&2
    export PENSUSE_LIVE_PASSWORD
fi

if [ "$need_sudo" -eq 1 ]; then
    exec sudo --preserve-env=PENSUSE_LIVE_PASSWORD make kiwi-iso
fi
exec make kiwi-iso
