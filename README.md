# PenSUSE

[![CI](https://github.com/Is0cre/PenSUSE/actions/workflows/ci.yml/badge.svg)](https://github.com/Is0cre/PenSUSE/actions/workflows/ci.yml)

**Security work with provenance.**

PenSUSE is an openSUSE-based security engineering and digital forensics workstation designed around reproducible execution, scoped operations, forensic provenance, isolated tooling, and evidence-backed analysis.

PenSUSE is not intended to be another security distribution containing thousands of preinstalled offensive tools.

Instead, it provides an operating environment for conducting professional security work from engagement creation through acquisition, analysis, findings, and reporting.

## Status

PenSUSE is currently in early architectural development.

Current milestone:

**M1 — Workbook Foundation (in progress)**

Do not consider the current system production-ready for forensic or security engagements.

---

# Design goals

PenSUSE is intended to support:

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

The initial PenSUSE platform targets:

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

---

# Workbooks

The central PenSUSE abstraction is a **workbook**.

A workbook represents one authorized:

- security engagement
- incident
- forensic investigation
- research case
- laboratory exercise

Example:

    pensuse workbook create client-2026
    pensuse workbook open client-2026

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

---

# Scope awareness

A workbook can declare permitted and excluded targets.

Examples:

- CIDRs
- individual hosts
- domains
- applications
- exclusions

PenSUSE-managed execution may warn when a target appears outside declared scope.

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

PenSUSE models evidence as structured objects rather than merely files.

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

PenSUSE intends to support local-first LLM-assisted investigation.

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

PenSUSE uses curated capability profiles rather than one enormous default installation.

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

PenSUSE should manage significant host changes using:

    SNAPSHOT
    APPLY
    VERIFY

Btrfs and Snapper provide recovery capability.

Case data must remain separate from operating-system rollback semantics.

---

# Technology sovereignty

PenSUSE should remain usable without mandatory external services.

Core functionality must not require:

- a cloud account
- an external control plane
- telemetry
- a hosted LLM
- a third-party SaaS product

Local mirrors, local OCI registries, local LLMs, and disconnected forensic laboratories should remain first-class deployment models.

PenSUSE is intended to be an independent European security engineering platform built on the openSUSE ecosystem.

---

# Non-goals

PenSUSE is not intended to:

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
pensuse workbook create client-2026
pensuse scope add client-2026 10.20.30.0/24
pensuse evidence import client-2026 disk.E01
pensuse run client-2026 --target 10.20.30.40 -- printf 'local check'
pensuse evidence list client-2026
pensuse run list client-2026
```

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

PenSUSE aims to preserve the context around those tools:

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

**PenSUSE helps operators conduct security work.**
