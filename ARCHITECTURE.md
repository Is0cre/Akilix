# Akilix Architecture

## 1. Purpose

This document defines the initial architectural boundaries of Akilix.

Architecture should evolve through explicit decisions rather than accidental implementation.

---

# 2. System model

Akilix consists conceptually of six layers.

    ┌─────────────────────────────────────────────┐
    │               Operator UX                   │
    │ CLI / Zsh / TUI / GUI / Reporting           │
    ├─────────────────────────────────────────────┤
    │              Workbook Engine                │
    │ Scope / Evidence / Provenance / Findings    │
    ├─────────────────────────────────────────────┤
    │          Security Execution Layer           │
    │ OCI tools / Native tools / Analysis VMs     │
    ├─────────────────────────────────────────────┤
    │          Forensic Acquisition Layer         │
    │ Block / Memory / Firmware / Device metadata │
    ├─────────────────────────────────────────────┤
    │             Akilix Platform                │
    │ Profiles / Policy / Snapshots / Privilege   │
    ├─────────────────────────────────────────────┤
    │             openSUSE Leap                   │
    │ RPM / Btrfs / SELinux / systemd / Podman    │
    └─────────────────────────────────────────────┘

The workbook engine is the defining Akilix component.

---

# 3. Host architecture

Initial target:

- openSUSE Leap 16.x
- x86_64
- systemd
- SELinux
- Btrfs
- Snapper
- Podman
- RPM
- zypper/libzypp

The host should remain conservative.

Applications with fast-moving dependencies should generally be isolated from the host using OCI environments.

Native installation remains appropriate for:

- operating-system administration
- hardware interaction
- filesystem tooling
- Podman
- virtualization
- trusted Akilix components
- acquisition utilities requiring direct hardware access

---

# 4. Filesystem model

System state and workbook state must not share rollback semantics.

Conceptually:

    /
      operating-system state

    /var/lib/akilix/
      Akilix machine state

    /srv/akilix/workbooks/
      workbook data

Exact layout may evolve during M0.

Workbooks should ideally support individual Btrfs subvolumes.

Btrfs snapshots provide workflow convenience and recovery.

Cryptographic hashes provide evidence integrity.

---

# 5. Workbook architecture

Each workbook has:

- immutable ID
- human name
- metadata
- scope configuration
- logging configuration
- evidence store
- artifact store
- provenance store
- notes
- findings
- timeline
- report material

Conceptual layout:

    workbook/
    ├── workbook.yaml
    ├── scope.yaml
    ├── evidence/
    │   ├── original/
    │   ├── acquired/
    │   └── manifests/
    ├── artifacts/
    │   ├── imported/
    │   ├── extracted/
    │   └── derived/
    ├── captures/
    ├── tool-output/
    ├── notes/
    ├── findings/
    ├── timeline/
    ├── reports/
    ├── logs/
    └── .akilix/
        ├── index.sqlite
        ├── manifest.jsonl
        ├── state.json
        └── locks/

Human-readable and machine-readable data should coexist.

The workbook must remain interpretable if `index.sqlite` is lost.

---

# 6. Object model

Initial object families:

## Workbook

Top-level engagement/case object.

## Evidence

Original or acquired material representing a source of forensic/security information.

## Artifact

A file, record, export, extraction, or derivative produced from evidence or a tool invocation.

## Invocation

A recorded execution event.

## Observation

A direct statement about something observed in evidence.

## TimelineEvent

A normalized temporal event.

## Claim

A factual, inferential, or hypothetical statement.

## Finding

A security-relevant conclusion supported by observations and artifacts.

## ReportSection

Presentation material generated from case objects.

---

# 7. Provenance graph

Akilix should support relationships such as:

    EV-0001
    disk-image.E01
         |
         | analyzed by
         v
    INV-0022
    sleuthkit
         |
         | produced
         v
    AR-0182
    filesystem-timeline.csv
         |
         | contains
         v
    OBS-0098
    SSH key created
         |
         | supports
         v
    FIND-0011
    Unauthorized SSH persistence

This graph is more important than terminal-session history.

---

# 8. Scope engine

Scope must support:

- IPv4
- IPv6
- CIDR
- hostnames
- domains
- wildcard domains
- URLs/applications
- explicit exclusions

Scope evaluation should produce structured results.

Example:

    ALLOW
    DENY
    UNKNOWN
    OVERRIDE

The execution engine may refuse an obvious out-of-scope target unless an explicit override is requested.

Overrides must be logged.

Scope checking is assistance.

It is not a containment boundary.

---

# 9. Execution engine

Managed execution should not simply call an opaque shell string.

Use explicit argument vectors.

Conceptually:

    akilix run \
        --environment network \
        -- nmap -sV 10.20.30.40

Execution flow:

    resolve workbook
        ↓
    inspect arguments
        ↓
    scope assessment
        ↓
    resolve OCI image digest
        ↓
    prepare mounts
        ↓
    execute
        ↓
    capture metadata
        ↓
    store outputs
        ↓
    identify generated artifacts
        ↓
    commit invocation record

Failed invocation logging must remain distinguishable from successful execution.

---

# 10. OCI model

Podman is the initial runtime.

Default properties:

- rootless
- no privileged mode
- no host PID namespace
- no arbitrary device access
- no mutable image identity
- minimal workbook mounts

Possible environments:

    akilix/recon
    akilix/network
    akilix/web
    akilix/directory
    akilix/cloud
    akilix/reverse
    akilix/malware
    akilix/forensics

These are conceptual environments, not mandatory one-tool-per-container designs.

---

# 11. Acquisition model

Acquisition is deliberately separate from ordinary analysis.

Three phases:

    ACQUIRE
    ANALYZE
    REPORT

Acquisition may interact with:

- `/dev`
- block layers
- storage controllers
- USB
- firmware interfaces
- kernel facilities

Therefore acquisition must have a narrowly designed privilege model.

Generic `podman --privileged` is not the Akilix acquisition architecture.

---

# 12. Privilege model

Initial principle:

The `akilix` command runs unprivileged.

Privileged operations should use explicit elevation.

If a privileged helper is eventually created, it must expose narrowly scoped operations rather than arbitrary command execution.

The project should avoid a privileged daemon until requirements prove it necessary.

---

# 13. LLM architecture

LLM functionality should communicate through an evidence/context broker.

Conceptually:

    Query
      |
      v
    Evidence Broker
      |
      +---- artifact index
      |
      +---- timeline index
      |
      +---- findings
      |
      v
    Context Builder
      |
      +---- LOCAL
      |
      +---- REMOTE-REDACTED
      |
      +---- REMOTE-EXPLICIT

The model should not simply receive unrestricted filesystem access.

Results should be represented structurally.

Example claim properties:

- classification
- statement
- evidence references
- model identity
- generation timestamp
- confidence where meaningful

---

# 14. API architecture

Initial releases do not require a daemon.

Core Go packages should expose reusable internal APIs.

Interfaces should eventually support:

- CLI
- Zsh integration
- TUI
- GUI
- possible local API

Business logic belongs in shared packages.

---

# 15. Image architecture

Image generation should be declarative.

Desired pipeline:

    Git
     |
     +-- Go source
     +-- RPM specs
     +-- profiles
     +-- OCI definitions
     +-- KIWI definitions
     +-- tests
     |
     v
    Build system
     |
     +-- Akilix RPM repository
     +-- OCI images
     +-- VM image
     +-- ISO
     +-- SBOM
     +-- checksums

The build process should not depend on hand-modified machines.

---

# 16. Repository strategy

Prefer upstream openSUSE packages whenever possible.

Akilix-specific packages should live separately.

Conceptual progression:

    Akilix:Factory
        |
        +---- Akilix:Testing
        |
        +---- Akilix:Stable

Avoid unnecessary forks.

A Akilix fork creates long-term maintenance responsibility and should require justification.

---

# 17. Offline and sovereignty model

Akilix core workflows should remain usable in disconnected environments.

Architecture must not require:

- external SaaS
- cloud login
- hosted orchestration
- third-party telemetry
- remote LLM access

Local RPM mirrors, OCI registries, model stores, documentation, and update channels should be supportable.

External integrations should remain adapters around the core platform.

---

# 18. Failure model

Akilix must assume operations can fail because of:

- process crashes
- power loss
- disk full
- filesystem error
- tool failure
- malformed evidence
- malformed manifests
- interrupted hashing
- broken container images
- missing OCI registry
- corrupted SQLite indexes

Evidence and provenance design must remain understandable after partial failure.

Generated records should use atomic writes where practical.

---

# 19. Security boundary philosophy

Akilix should clearly distinguish:

## Security boundary

A control intended to prevent unauthorized action.

Examples:

- SELinux
- Unix permissions
- mount flags
- namespace isolation

## Safety guard

A control designed to help the operator avoid mistakes.

Examples:

- scope warnings
- acquisition warnings
- profile-change previews

Do not describe safety guards as hard security boundaries.

---

# 20. Architectural rule

Before introducing a major component, ask:

1. Which workbook object does this operate on?
2. What provenance does it generate?
3. What privileges does it need?
4. What network access does it need?
5. Can it modify original evidence?
6. Can it be reproduced?
7. How can the operator inspect what it did?
8. How can failure be detected?
9. How is rollback handled?
10. Why does this belong on the host rather than in an isolated environment?

If these questions cannot be answered, the component is not ready for integration.
