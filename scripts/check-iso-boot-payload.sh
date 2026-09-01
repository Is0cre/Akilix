#!/bin/sh
set -eu

if [ "$#" -ne 1 ] || [ ! -f "$1" ]; then
	echo "usage: $0 ISO_PATH" >&2
	exit 2
fi
for tool in xorriso lsinitrd file; do
	command -v "$tool" >/dev/null 2>&1 || {
		echo "required boot-payload inspection tool missing: $tool" >&2
		exit 1
	}
done

iso=$1
check_root=$(mktemp -d /tmp/akilix-iso-payload.XXXXXX)
trap 'rm -rf -- "$check_root"' EXIT

xorriso -osirrox on -indev "$iso" \
	-extract /boot/x86_64/loader/linux "$check_root/linux" \
	-extract /boot/x86_64/loader/initrd "$check_root/initrd" \
	>/dev/null 2>&1

kernel_description=$(file "$check_root/linux")
kernel_release=$(printf '%s\n' "$kernel_description" | sed -n 's/.*version \([^ ]*\).*/\1/p')
if [ -z "$kernel_release" ]; then
	echo "unable to identify ISO kernel release" >&2
	exit 1
fi

listing=$check_root/initrd.list
lsinitrd "$check_root/initrd" >"$listing"
grep -Fq "usr/lib/modules/$kernel_release/" "$listing" || {
	echo "ISO kernel $kernel_release has no matching initrd module tree" >&2
	exit 1
}
grep -Eq '^lrwxrwxrwx .* init -> usr/lib/systemd/systemd$' "$listing" || {
	echo "initrd /init does not resolve to systemd" >&2
	exit 1
}

lsinitrd -f usr/lib/systemd/systemd "$check_root/initrd" >"$check_root/systemd"
lsinitrd -f usr/lib64/ld-linux-x86-64.so.2 "$check_root/initrd" >"$check_root/loader"
for payload in systemd loader; do
	description=$(file "$check_root/$payload")
	printf '%s\n' "$description" | grep -Fq 'ELF 64-bit' || {
		echo "initrd $payload is not a 64-bit ELF payload: $description" >&2
		exit 1
	}
	printf '%s\n' "$description" | grep -Fq 'x86-64' || {
		echo "initrd $payload is not x86-64: $description" >&2
		exit 1
	}
done

printf 'ISO boot payload valid: kernel=%s init=systemd/x86-64\n' "$kernel_release"
