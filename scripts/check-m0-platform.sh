#!/bin/sh
set -u

failures=0

pass() { printf '%s\n' "PASS: $*"; }
warn() { printf '%s\n' "WARN: $*"; }
fail() { printf '%s\n' "FAIL: $*"; failures=$((failures + 1)); }

if command -v pensuse >/dev/null 2>&1; then
    pass "pensuse CLI found at $(command -v pensuse)"
    pensuse version >/dev/null 2>&1 && pass "pensuse version" || fail "pensuse version"
else
    fail "pensuse CLI not found in PATH"
fi

if command -v btrfs >/dev/null 2>&1; then
    pass "btrfs utility available"
    btrfs filesystem show / >/dev/null 2>&1 && pass "root filesystem inspected" || warn "root filesystem is not Btrfs or is unavailable"
else
    fail "btrfs utility not found"
fi

if command -v snapper >/dev/null 2>&1; then
    pass "snapper utility available"
    snapper list-configs >/dev/null 2>&1 && pass "snapper configurations readable" || warn "no readable Snapper configuration"
else
    fail "snapper utility not found"
fi

if command -v podman >/dev/null 2>&1; then
    pass "podman utility available"
    podman info --format '{{.Host.Security.Rootless}}' >/dev/null 2>&1 && pass "podman info available" || fail "podman info"
    if [ "$(id -u)" -eq 0 ]; then
        warn "running as root; rootless Podman requires a regular operator account"
    else
        podman unshare true >/dev/null 2>&1 && pass "rootless user namespace" || fail "rootless user namespace"
    fi
else
    fail "podman utility not found"
fi

if command -v getenforce >/dev/null 2>&1; then
    pass "SELinux mode: $(getenforce)"
elif command -v sestatus >/dev/null 2>&1; then
    pass "SELinux status available"
else
    warn "SELinux status utility not available"
fi

if command -v ss >/dev/null 2>&1; then
    pass "listener inspection available"
    ss -lnt
else
    fail "ss listener inspection utility not found"
fi

if [ "$failures" -ne 0 ]; then
    printf '%s\n' "$failures required M0 check(s) failed" >&2
    exit 1
fi

pass "M0 platform checks completed"
