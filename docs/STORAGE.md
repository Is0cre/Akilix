# Storage and network filesystem interoperability

The Akilix live image includes:

- `ntfs-3g` for NTFS interoperability alongside the kernel's native support;
- `cifs-utils` and `samba-client` for operator-initiated SMB access;
- Samba server binaries for explicit laboratory workflows;
- OpenZFS user space, `zfs-kmp-default`, and its UEFI certificate from the
  openSUSE `filesystems/16.0` OBS repository.

Support does not imply automatic access. SMB server services and ZFS import,
mount, share, and event services are disabled by the Akilix systemd preset.
The image does not automatically connect to SMB servers, publish shares,
import ZFS pools, or mount discovered filesystems.

ZFS is an experimental host component. It is an out-of-tree kernel module and
must match the running Leap kernel. Secure Boot systems may require explicit
operator enrollment of the repository's module-signing certificate. Operators
must verify `modinfo zfs` and the running kernel before relying on it.

Forensic source media should be protected and inventoried before filesystem
tools are used. Software read-only state is not a substitute for a hardware
write blocker. Importing or mounting a filesystem is an explicit operator
action and should be recorded in the active workbook.

## Login message

`/etc/motd` provides the static Akilix safety banner. Interactive shells add a
short colored local status line. It runs no network command. Blinking of the
acquisition warning is available only when explicitly requested:

    export AKILIX_MOTD_BLINK=1

Blinking is disabled by default for accessibility and terminal compatibility.
