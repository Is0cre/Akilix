# AGENTS.md — PenSUSE Engineering Instructions

## Project

PenSUSE is an openSUSE-based professional security engineering and digital forensics workstation.

PenSUSE is NOT intended to be a Kali Linux or Parrot OS clone.

The project is not measured by how many offensive-security tools it contains.

PenSUSE is designed around:

- reproducible security work
- scoped engagements
- forensic evidence integrity
- provenance
- disposable tool environments
- conservative host-level acquisition
- strong Linux system administration
- transactional host changes
- transparent logging
- local-first analysis
- traceable LLM-assisted investigation

The fundamental PenSUSE abstraction is the **workbook**.

A workbook represents one authorized engagement, investigation, incident, case, or laboratory exercise.

---

# Engineering priorities

When making architectural decisions, prefer these properties in this order:

1. Evidence safety
2. Reproducibility
3. Explicit operator control
4. Provenance
5. Isolation
6. Auditability
7. Recoverability
8. Simplicity
9. Usability
10. Feature count

Do not sacrifice evidence integrity or provenance for convenience.

---

# Distribution foundation

Initial PenSUSE development targets:

- openSUSE Leap 16.x
- x86_64
- RPM packaging
- libzypp/zypper
- Btrfs
- Snapper
- systemd
- SELinux
- Podman
- rootless containers where possible
- KIWI NG for image generation
- Agama-compatible installation direction
- Open Build Service for package infrastructure where appropriate

Do not replace native openSUSE infrastructure without a strong technical reason.

---

# Language

The primary PenSUSE CLI/backend should be implemented in Go unless an architectural decision record explicitly changes this.

Shell scripting is allowed for:

- build glue
- test harnesses
- installation helpers
- CI integration

Shell scripts must not become the primary application architecture.

Python may be used where an ecosystem dependency strongly justifies it, particularly in forensic or analytical tooling, but PenSUSE core functionality must not depend on arbitrary Python environments.

---

# Core command

The primary interface is:

    pensuse

Examples:

    pensuse version

    pensuse workbook create client-2026
    pensuse workbook open client-2026
    pensuse workbook status

    pensuse scope check 10.20.30.40

    pensuse evidence import disk.E01
    pensuse evidence verify EV-000001

    pensuse run ...

    pensuse acquire ...

    pensuse profile ...

The CLI, future TUI, future GUI, and shell integration must reuse common backend logic.

Do not implement separate business logic for each interface.

---

# Architecture rule

PenSUSE consists conceptually of these layers:

1. Operator UX
2. Workbook engine
3. Security execution layer
4. Forensic acquisition layer
5. PenSUSE platform management
6. openSUSE host

The workbook engine is the central layer.

Security tools are consumers of workbook services.

Security tools are NOT the core architecture.

---

# Workbooks

Every workbook must have:

- a human-readable name
- an immutable unique ID
- creation timestamp
- declared scope
- logging policy
- evidence storage
- artifact storage
- invocation history
- provenance information
- findings storage
- report material

Prefer UUIDv7 for globally unique workbook and object identifiers.

A workbook should remain interpretable from its files even if the PenSUSE application or local index database is unavailable.

SQLite may be used as an index.

SQLite must NOT become the sole canonical storage for forensic evidence provenance.

---

# Evidence

Original evidence is sacred.

Analysis operations must never modify original evidence.

Original evidence should be stored separately from:

- derived evidence
- extracted artifacts
- tool output
- working files
- reports
- temporary files

All imported or acquired evidence must support cryptographic hashing.

SHA-256 is the initial mandatory evidence hash.

Additional hashes may be recorded where useful.

Btrfs snapshots are NOT evidence integrity mechanisms.

Snapshots provide operational recovery.

Cryptographic hashes and provenance provide evidence integrity.

---

# Evidence objects

Do not model the system as a collection of filenames.

Model important objects explicitly.

Initial conceptual object types include:

- Workbook
- Evidence
- Artifact
- Invocation
- Observation
- TimelineEvent
- Claim
- Finding
- ReportSection

Relationships between these objects must be representable.

Example:

    Evidence
        |
        | parsed by
        v
    Invocation
        |
        | generated
        v
    Artifact
        |
        | supports
        v
    Observation
        |
        | supports
        v
    Finding

---

# Provenance

Every PenSUSE-managed tool invocation that can generate security-relevant output should be capable of recording:

- invocation ID
- workbook ID
- start timestamp
- end timestamp
- executor type
- tool name
- tool version
- container image name where applicable
- immutable container image digest
- exact argument vector
- environment information where appropriate
- scope decision
- scope override where applicable
- exit code
- stdout artifact
- stderr artifact
- generated files
- input evidence references
- output artifact references

Do not depend on parsing shell history to reconstruct provenance.

---

# Scope

PenSUSE assists operators with engagement scope.

Scope may include:

- CIDRs
- individual IP addresses
- hostnames
- domains
- wildcard domains
- URLs/applications
- exclusions

PenSUSE may warn or block PenSUSE-managed execution when obvious scope violations are detected.

PenSUSE scope handling MUST NOT be described as a security sandbox.

Security tools may behave in ways PenSUSE cannot completely inspect or constrain.

Scope enforcement is operator assistance, not perfect containment.

Explicit scope overrides should be recorded.

---

# Containers

Podman is the default container runtime.

Prefer rootless containers.

Security-tool environments should be reproducible and disposable.

Containers must not own workbook data.

Workbook data lives outside containers.

Container mounts should distinguish:

- original evidence — read-only
- derived evidence — controlled write
- tool output — writable
- temporary workspace — writable

OCI image digests must be recorded for managed invocations.

Do not rely only on mutable tags such as `latest`.

Do not require one container per individual tool.

Prefer coherent tool-family environments such as:

- recon
- network
- web
- directory
- cloud
- reverse
- malware
- forensics

Do not automatically give containers:

- host networking
- privileged mode
- host PID namespace
- unrestricted device access

Any privilege expansion must be explicit.

---

# Host-level acquisition

Raw forensic acquisition is a separate subsystem from ordinary container execution.

Do not hide hardware acquisition behind generic privileged containers.

Conceptually separate:

    ACQUIRE
    ANALYZE
    REPORT

Host-level acquisition may require explicit privilege elevation.

Acquisition should use the smallest practical privileged surface.

Potential acquisition targets include:

- block devices
- physical disks
- removable media
- disk images
- SMART metadata
- NVMe metadata
- partition tables
- USB metadata
- firmware/UEFI information
- filesystem metadata
- memory where technically supported

Evidence source devices should default toward read-only operation wherever technically possible.

Software read-only protection must never be presented as equivalent to a hardware forensic write blocker.

---

# Privilege

Default operation should be unprivileged.

Do not create a permanently privileged daemon during early milestones.

If privileged operations become necessary:

- isolate them
- minimize accepted input
- minimize code surface
- authenticate callers
- log requested actions
- use explicit elevation
- maintain clear privilege boundaries

Do not solve permission problems by making the main PenSUSE process run as root.

---

# Host modifications

PenSUSE host changes should follow:

    SNAPSHOT
    APPLY
    VERIFY

Where technically applicable.

Profile installation and major configuration changes should integrate with Btrfs/Snapper.

Rollback must remain easy.

Host state and case/workbook data must not share rollback semantics.

Rolling back the operating system must not silently roll back investigative evidence.

---

# Profiles

PenSUSE uses curated profiles.

Initial conceptual profiles:

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

Profiles may contain combinations of:

- RPM dependencies
- configuration
- OCI image definitions
- shell completions
- policy
- integration metadata

Profiles should not simply install enormous tool collections on the host.

---

# LLM architecture

LLM integration is a major PenSUSE capability but is NOT part of M0.

Supported conceptual modes:

- OFF
- LOCAL
- REMOTE-REDACTED
- REMOTE-EXPLICIT

Sensitive investigations should naturally favor LOCAL.

LLMs must not receive arbitrary unrestricted filesystem access.

Prefer an evidence-broker architecture where models request specific indexed artifacts or records.

LLM-generated material must distinguish:

- FACT
- INFERENCE
- HYPOTHESIS

LLM output is never automatically evidence.

A factual assertion presented by an LLM must be traceable to source records whenever possible.

Remote transmission of raw evidence must require explicit operator action.

Do not implement hidden automatic upload.

---

# Logging

No hidden monitoring.

The operator must be able to determine exactly which logging features are active.

Logging capabilities may include:

- command metadata
- command arguments
- container metadata
- evidence hashing
- stdout capture
- stderr capture
- generated-file tracking
- packet metadata
- terminal recording

Features such as terminal recording or packet capture must not silently activate.

No telemetry by default.

No secret call-home functionality.

---

# Network behavior

PenSUSE must not:

- automatically scan networks
- automatically attack targets
- automatically open listeners
- automatically expose services
- silently upload workbook data
- silently contact telemetry services

Opening a workbook must not initiate network activity.

Security actions require explicit operator actions.

---

# Technology sovereignty

Core PenSUSE functionality must not require:

- a cloud account
- a hosted control plane
- an external SaaS service
- a telemetry endpoint
- a US or non-European service provider

Offline-capable operation is a design objective.

Local mirrors and self-hosted infrastructure should be supported.

Remote services are optional integrations, not architectural dependencies.

---

# Security tooling

Do NOT begin development by adding hundreds of security tools.

Do NOT add a tool merely because Kali contains it.

A tool should be added because a PenSUSE workflow requires it.

Before adding significant tooling, the following architecture must exist:

- workbook model
- evidence model
- provenance model
- scope model
- execution model

Tool count is not a project KPI.

---

# M0 restrictions

During M0, do NOT implement:

- offensive security suites
- Metasploit integration
- vulnerability scanners
- exploit frameworks
- malware detonation
- Wi-Fi attacks
- AD attack tooling
- LLM integration
- GUI
- remote services
- privileged background daemons
- large security package profiles

M0 establishes the platform.

---

# M0 goals

M0 should establish:

- repository structure
- architecture documents
- PenSUSE invariants
- Go module
- minimal `pensuse` CLI
- `pensuse version`
- basic configuration handling
- RPM packaging skeleton
- KIWI image definition
- Btrfs layout
- Snapper verification
- rootless Podman verification
- minimal Leap 16 bootable image
- automated tests
- CI entry points

The M0 image should intentionally contain very few security tools.

---

# M1 restrictions

M1 focuses on:

- workbooks
- scope
- evidence importing
- evidence hashing
- provenance
- managed execution

Do not expand M1 into a general security-tool distribution.

---

# Coding standards

Prefer:

- clear Go packages
- typed structures
- explicit errors
- deterministic serialization
- atomic file writes
- fsync where evidence durability requires it
- secure default permissions
- structured logs
- testable dependency boundaries
- minimal hidden global state

Avoid:

- giant files
- giant command handlers
- hidden side effects
- implicit network operations
- arbitrary shell command construction
- parsing human-readable command output when a structured interface exists
- mutable global configuration
- unnecessary dependencies

Use `exec.CommandContext` style argument vectors rather than shell command strings wherever possible.

Never concatenate untrusted input into shell commands.

---

# Filesystem safety

Code handling evidence must consider:

- symlinks
- hard links
- path traversal
- mount boundaries
- race conditions
- partial writes
- interrupted operations
- permissions
- ownership
- filesystem-full conditions

Do not assume filenames are trusted.

Do not silently overwrite evidence.

---

# Serialization

Human-controlled configuration should prefer YAML where appropriate.

Machine-generated records should prefer deterministic JSON or JSONL.

Schemas should be versioned.

Example:

    pensuse.workbook.v1
    pensuse.evidence.v1
    pensuse.invocation.v1

Schema evolution must be deliberate.

---

# Tests

Tests are mandatory.

Initial classes:

- unit tests
- integration tests
- VM/image tests
- forensic fixture tests
- failure-path tests
- reproducibility tests

Important behaviors should be tested as properties/invariants where possible.

Examples:

- evidence import never overwrites an existing original
- analysis mounts originals read-only
- opening workbook causes no network activity
- OCI invocation records immutable digest
- failed operations do not leave valid-looking partial manifests
- corrupted indexes do not destroy canonical evidence metadata

---

# Development behavior for agents

Before implementing a milestone:

1. Read AGENTS.md.
2. Read ARCHITECTURE.md.
3. Read docs/INVARIANTS.md.
4. Read the active milestone document.
5. Inspect existing implementation.
6. Identify conflicts between requested changes and project invariants.
7. Prefer the smallest implementation satisfying the milestone.
8. Add tests.
9. Run relevant tests.
10. Update documentation when architecture changes.

Do not silently reinterpret project architecture.

If implementation reveals that an architectural assumption is wrong, document the issue before replacing the design.

---

# Definition of done

A change is not complete merely because it compiles.

A change is complete when:

- functionality works
- tests cover success and important failure paths
- security implications were considered
- evidence/provenance invariants remain satisfied
- user-visible behavior is documented
- no hidden network activity was introduced
- no unnecessary privileges were introduced
- error states are understandable
- relevant tests pass

---

# Project identity

PenSUSE helps operators conduct security work.

It is not merely a system that happens to contain security tools.

The design should always preserve that distinction.
