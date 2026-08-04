#!/bin/sh
set -eu

# Read-only baseline check. It does not alter services or network state.
if command -v ss >/dev/null 2>&1; then
    # Port/state inspection is available to ordinary users; process ownership
    # details may require privilege and are intentionally not requested here.
    ss -lnt
elif command -v netstat >/dev/null 2>&1; then
    netstat -lntup
else
    echo "no socket inspection tool available" >&2
    exit 1
fi
