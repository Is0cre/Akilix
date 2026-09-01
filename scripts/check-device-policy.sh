#!/bin/sh
set -eu

rule=image/kiwi-iso/root/etc/udev/rules.d/99-akilix-forensic-block.rules
tmpfiles=image/kiwi-iso/root/usr/lib/tmpfiles.d/akilix-device-events.conf

test -f "$rule"
test -f "$tmpfiles"
grep -F 'SUBSYSTEM=="block"' "$rule" >/dev/null
grep -F 'ENV{DEVTYPE}=="disk"' "$rule" >/dev/null
grep -F 'ATTR{ro}="1"' "$rule" >/dev/null
grep -F 'ENV{UDISKS_IGNORE}="1"' "$rule" >/dev/null
grep -F 'akilix-udev-handler add' "$rule" >/dev/null
grep -F 'akilix-udev-handler remove' "$rule" >/dev/null
if grep -E 'SUBSYSTEM=="usb"|KERNEL=="event|KERNEL=="mouse|KERNEL=="hid' "$rule" >/dev/null; then
    echo "device policy unexpectedly targets non-storage USB input" >&2
    exit 1
fi
grep -F '/run/akilix/device-events 0770 root wheel' "$tmpfiles" >/dev/null
printf '%s\n' 'device policy check: external whole-disk storage is queued read-only'
