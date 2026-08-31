# Akilix

> The project was renamed to Akilix during pre-release development. See
> [the identity migration note](docs/IDENTITY.md) for storage, repository, and
> visual-branding boundaries.

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="branding/web/akilix-horizontal-dark.png">
    <source media="(prefers-color-scheme: light)" srcset="branding/web/akilix-horizontal-light.png">
    <img alt="Akilix — Security work with provenance" src="branding/web/akilix-horizontal-light.png" width="720">
  </picture>
</p>

[![CI](https://github.com/Is0cre/Akilix/actions/workflows/ci.yml/badge.svg)](https://github.com/Is0cre/Akilix/actions/workflows/ci.yml)

**Security work with provenance.**

Akilix is an openSUSE-based security engineering and digital forensics workstation designed around reproducible execution, scoped operations, forensic provenance, isolated tooling, and evidence-backed analysis.

Akilix is not intended to be another security distribution containing thousands of preinstalled offensive tools.

Instead, it provides an operating environment for conducting professional security work from engagement creation through acquisition, analysis, findings, and reporting.

## Status

Akilix is currently in early architectural development.

Current status:

**M0 platform image validated; M1/M2 implementation and M3 profile metadata in progress**

The Leap 16 live image boots in both BIOS and UEFI modes. Runtime hardware
validation (Btrfs, Snapper, SELinux, and rootless Podman on the target laptop)
is still pending.

Do not consider the current system production-ready for forensic or security engagements.

---

# Design goals

Akilix is intended to support:

- penetration testing
- red-team operations
- blue-team investigation
- incident response
- digital forensics
- reverse engineering
- malware analysis
- network analysis
- wireless security
- cloud security
- container security
- hardware and disk investigation

The system is built around several principles:

- explicit operator action
- reproducible tool environments
- local-first case data
- evidence preservation
- cryptographic hashing
- traceable provenance
- scope awareness
- transparent logging
- transactional host changes
- strong privilege boundaries
- no telemetry by default
- no mandatory external service
- offline-capable operation where practical

---

# Foundation

The initial Akilix platform targets:

- openSUSE Leap 16.x
- RPM
- libzypp
- zypper
- Btrfs
- Snapper
- SELinux
- systemd
- Podman
- Open Build Service
- KIWI NG
- Agama-compatible installation

The host is intended to remain relatively conservative.

Fast-moving security applications should generally run in reproducible OCI environments rather than continuously changing the base operating system.

The intended graphical environment is a lightweight, keyboard-first Sway
session with tiling enabled by default. The console remains a supported
recovery path. The development image uses an explicitly approved openSUSE OBS
source with pinned, offline-auditable package identities; graphical auto-login
remains disabled until runtime validation is complete.

---

# Workbooks

The central Akilix abstraction is a **workbook**.

A workbook represents one authorized:

- security engagement
- incident
- forensic investigation
- research case
- laboratory exercise

Example:

    akilix workbook create client-2026
    akilix workbook open client-2026

A workbook may contain:

- scope
- evidence
- acquisition records
- artifacts
- captures
- notes
- tool output
- command records
- hashes
- logs
- findings
- timeline events
- report material

## Passive hardware inspection

The first forensic-acquisition command inventories whole disks and their
partitions without opening raw device nodes or changing device state:

    akilix acquire inspect
    akilix acquire inspect --json
    akilix acquire record client-2026
    sudo --preserve-env=AKILIX_WORKBOOK_ROOT akilix acquire protect client-2026 /dev/sdb
    akilix device trust add /dev/sdb "known lab disk"
    akilix device trust list

It reports capacity, transport, vendor/model/serial/WWN identifiers,
filesystems, UUIDs, mount state, and current kernel read-only state. System
disks are identified conservatively through root and boot mount ancestry.
`acquisition_candidate` means only that a disk was not identified as the
running system disk; it is not authorization and does not prove the device is
external. See [M4.md](M4.md).

The explicit `protect` operation refuses system disks, mounted disks, and
partition paths. It records a durable request before invoking `blockdev
--setro`, verifies the resulting kernel state, and then records an immutable
applied or failed event. Akilix never invokes `sudo` itself and provides no
automatic write-enable command. Kernel software read-only state is an
operational safeguard, not a hardware forensic write blocker.

Trusted-device policy uses a stable WWN, or a serial combined with vendor and
model, rather than transient `/dev` names. Devices without a sufficiently
stable identity cannot be added. Trust entries and revocations are retained in
the local state registry. Trust is descriptive policy only: it never makes
storage writable, bypasses protection, or constitutes acquisition authority.
The ISO also stages a checksum-pinned `usb.ids` snapshot for offline friendly
names; those unauthenticated VID/PID labels are never treated as identity.
For USB storage, numeric VID/PID values from local udev metadata are combined
with the serial when no WWN exists. Friendly database names remain display-only.

---

# Scope awareness

A workbook can declare permitted and excluded targets.

Examples:

- CIDRs
- individual hosts
- domains
- applications
- exclusions

Akilix-managed execution may warn when a target appears outside declared scope.

Scope assistance does not pretend to be perfect containment.

The operator remains responsible for authorization and tool behavior.

---

# Reproducible tooling

Security tool families may run through Podman environments.

Examples:

- recon
- network
- web
- directory
- cloud
- reverse
- malware
- forensics

Containers are intended to be disposable.

Workbook data persists outside them.

Managed invocations can record:

- tool
- tool version
- OCI image digest
- argument vector
- timestamps
- exit code
- stdout
- stderr
- generated files
- workbook ID
- provenance relationships

---

# Forensic architecture

Forensic workflow is divided conceptually into:

    ACQUIRE
        ↓
    ANALYZE
        ↓
    REPORT

Acquisition of physical evidence is intentionally separated from generic container execution.

Evidence devices should be handled conservatively and made read-only where technically practical.

Software read-only handling is not represented as equivalent to a hardware forensic write blocker.

Original evidence must not be modified by analysis operations.

---

# Evidence and provenance

Akilix models evidence as structured objects rather than merely files.

A finding should be capable of tracing back through:

    Finding
       ↓
    Observation
       ↓
    Artifact
       ↓
    Invocation
       ↓
    Evidence

Important operations may record:

- cryptographic hashes
- parser/tool identity
- tool version
- container digest
- command
- file
- record
- event
- byte offset
- provenance relationships

---

# LLM-assisted analysis

Akilix intends to support local-first LLM-assisted investigation.

Conceptual modes:

    OFF
    LOCAL
    REMOTE-REDACTED
    REMOTE-EXPLICIT

Possible capabilities include:

- log summarization
- timeline construction
- event correlation
- artifact clustering
- lead identification
- tool-output summarization
- notebook maintenance
- finding drafting
- report assistance

LLM output must distinguish:

- FACT
- INFERENCE
- HYPOTHESIS

LLM output never silently becomes evidence.

Claims should remain traceable to source material.

---

# Logging

Logging must be transparent.

Example future output:

    Command logging         enabled
    Container logging       enabled
    Evidence hashing        enabled
    stdout capture          enabled
    stderr capture          enabled
    Packet metadata         disabled
    Full terminal capture   disabled

No hidden operator monitoring is part of the design.

---

# Profiles

Akilix uses curated capability profiles rather than one enormous default installation.

Planned profiles include:

- CORE
- RECON
- NETWORK
- WEB
- WIRELESS
- DIRECTORY
- CLOUD
- REVERSE
- MALWARE
- FORENSICS
- INCIDENT-RESPONSE
- BLUE

Installing a profile does not necessarily mean installing all of its applications directly onto the host.

---

# Transactional host management

Akilix should manage significant host changes using:

    SNAPSHOT
    APPLY
    VERIFY

Btrfs and Snapper provide recovery capability.

Case data must remain separate from operating-system rollback semantics.

---

# Technology sovereignty

Akilix should remain usable without mandatory external services.

Core functionality must not require:

- a cloud account
- an external control plane
- telemetry
- a hosted LLM
- a third-party SaaS product

Local mirrors, local OCI registries, local LLMs, and disconnected forensic laboratories should remain first-class deployment models.

Akilix is intended to be an independent European security engineering platform built on the openSUSE ecosystem.

---

# Non-goals

Akilix is not intended to:

- maximize installed security-tool count
- automatically scan networks
- automatically attack systems
- silently open listeners
- silently upload case data
- perform hidden telemetry
- replace professional authorization procedures
- claim software-only write protection equals forensic hardware protection

---

# Development

The initial implementation and reproducible build instructions are in
[BUILD.md](BUILD.md). Host-specific M0 observations are recorded in
[docs/PLATFORM-BASELINE.md](docs/PLATFORM-BASELINE.md).
The container execution foundation is documented in [M2.md](M2.md).

Current local workflow examples:

```sh
akilix workbook create client-2026
akilix tui client-2026
akilix scope add client-2026 10.20.30.0/24
akilix evidence import client-2026 disk.E01
akilix run client-2026 --target 10.20.30.40 -- printf 'local check'
akilix evidence list client-2026
akilix run list client-2026
```

Shell completion is available for Zsh and Bash. Install it with
`sudo make install-completion`, then restart the shell or reload completion
initialization.

The project is currently establishing its architecture before adding substantial security tooling.

See:

- `AGENTS.md`
- `ARCHITECTURE.md`
- `ROADMAP.md`
- `M0.md`
- `docs/INVARIANTS.md`

for development rules and current goals.

---

# Project philosophy

Traditional security distributions largely help users obtain and run tools.

Akilix aims to preserve the context around those tools:

    authorization
         +
       scope
         +
      execution
         +
      evidence
         +
     provenance
         +
      analysis
         +
      findings
         +
      reporting

**Akilix helps operators conduct security work.**
