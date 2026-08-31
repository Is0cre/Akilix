#!/bin/sh
set -u

failures=0

pass() { printf '%s\n' "PASS: $*"; }
warn() { printf '%s\n' "WARN: $*"; }
fail() { printf '%s\n' "FAIL: $*"; failures=$((failures + 1)); }

if command -v akilix >/dev/null 2>&1; then
    pass "akilix CLI found at $(command -v akilix)"
    akilix version >/dev/null 2>&1 && pass "akilix version" || fail "akilix version"
else
    fail "akilix CLI not found in PATH"
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
    if [ "$(id -u)" -eq 0 ]; then
        warn "running as root; rootless Podman requires a regular operator account"
	    elif command -v akilix >/dev/null 2>&1; then
	        akilix container doctor >/dev/null 2>&1 && pass "Akilix rootless container runtime" || fail "Akilix rootless container runtime"
    else
	        podman info --format '{{.Host.Security.Rootless}}' >/dev/null 2>&1 && pass "podman info available" || fail "podman info"
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
