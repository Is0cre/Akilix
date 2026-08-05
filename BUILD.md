# PenSUSE M0 build guide

The M0 implementation is intentionally small and uses the Go standard library only.

## Local CLI

```sh
go build -o pensuse ./cmd/pensuse
./pensuse version
./pensuse version --json
go test ./...
```

The complete local validation sequence is also available as:

```sh
make check
```

In restricted build environments where the default home directory is
read-only, use writable temporary Go caches:

```sh
GOCACHE=/tmp/pensuse-go-cache GOPATH=/tmp/pensuse-gopath go test ./...
```

## RPM

On openSUSE Leap with `golang` and `rpmbuild` installed:

```sh
make rpm
```

The RPM target creates a source archive, builds with `rpmbuild`, and installs
the CLI at `/usr/bin/pensuse`.

The spec installs the CLI at `/usr/bin/pensuse`.

## Development image

With KIWI NG and an openSUSE Leap 16 repository available:

```sh
make kiwi
```

For a Ventoy/USB bootable ISO:

```sh
sudo env PENSUSE_LIVE_PASSWORD='choose-a-local-test-password' make kiwi-iso
```

KIWI image builds require root privileges and a writable build host.
Choose a local test password when building the live image; it is hashed into
the image and is not a repository default.

The image definition is under `image/kiwi/`. The image intentionally contains
only platform prerequisites; it does not install security-tool collections.

## Host verification

The read-only listener check is:

```sh
./scripts/check-network-baseline.sh
```

After booting the Leap 16 image, run the consolidated M0 platform check as a
regular operator account:

```sh
make vm-check
```

The same check is installed in the live ISO as `pensuse-m0-check`, so it can
be run directly from the console without a checkout.

The check is observational: it does not create snapshots, pull container
images, change SELinux policy, or start services. A root login is useful for
inspection, but rootless Podman must be checked from the intended unprivileged
operator account.

Snapper, Btrfs, SELinux, and rootless Podman checks require execution inside a
Leap 16 development image and are not silently simulated by the build.
