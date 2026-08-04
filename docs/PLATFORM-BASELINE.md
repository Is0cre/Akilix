# M0 Platform Baseline

This file records observations from the development host, rather than assuming
that the target image already behaves as designed.

## Observed host

- Distribution: openSUSE Leap 16.0
- Kernel: 6.12.0-160000.36-default
- Architecture: x86_64
- `zypper`: present

## Tooling not available during this run

The following commands were not installed in the current environment:

- `go`
- `rpmbuild`
- `kiwi-ng`
- `podman`

Therefore Go tests, RPM builds, KIWI image builds, and rootless-container
verification remain pending. They must be executed in a provisioned Leap 16
development environment; they are not simulated by the repository scripts.

The toolchain was subsequently provisioned: Go tests, `go vet`, and the RPM
build pass. Rootless Podman reports successfully with the `runc` runtime and
rootless mode enabled. No local images are currently present, so no image was
pulled and no network access was introduced for testing.

## Package-manager observation

An offline `zypper` package query could not initialize because RPM public-key
data under `/usr/lib/rpm/gnupg/keys` was unavailable. This should be repaired
in the build environment before installing M0 dependencies. No package or
system state was changed by this project run.

## Network baseline

The read-only listener inspection script is available at
`scripts/check-network-baseline.sh`. It does not start services, alter sockets,
or contact remote services.

## Privileged checks still pending

The KIWI image build correctly refuses to run without root privileges. The
current environment also lacks `btrfs` and `getenforce`, and Snapper cannot
access the system bus. These checks must run on the provisioned Leap 16 VM or
build host, with explicit administrative authorization.
