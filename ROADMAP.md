# Akilix Development Roadmap

This roadmap defines architectural ordering.

Milestones may gain sub-milestones.

Later milestones should not be implemented prematurely merely because a feature appears interesting.

---

# M0 — Architecture and Platform Foundation

Build the smallest correct Akilix system.

Primary subjects:

- openSUSE Leap base
- repository
- architecture
- invariants
- Go CLI
- RPM packaging
- KIWI image
- Btrfs
- Snapper
- SELinux baseline
- Podman
- build/test infrastructure

Security tooling is deliberately minimal.

---

# M1 — Workbook Foundation

Implementation is in progress: workbooks, evidence hashing, scope assistance,
and native provenance-recorded execution are available in the current CLI.

Implement the defining Akilix abstraction.

Primary subjects:

- workbook identity
- workbook creation
- workbook opening
- workbook state
- directory layout
- schema versioning
- scope configuration
- evidence importing
- evidence hashing
- evidence manifests
- invocation records
- basic managed execution

Expected commands include:

    akilix workbook create
    akilix workbook list
    akilix workbook open
    akilix workbook close
    akilix workbook status

    akilix scope add
    akilix scope remove
    akilix scope list
    akilix scope check

    akilix evidence import
    akilix evidence list
    akilix evidence verify

    akilix run

The initial tool used to demonstrate managed execution may be deliberately boring.

Tooling breadth is not an M1 goal.

---

# M2 — Reproducible Tool Execution

Initial digest-pinned container execution, hardened Podman argument generation,
and container provenance recording are implemented. Runtime image and VM
validation remain pending on the hardware test host.

Build mature execution infrastructure.

Primary subjects:

- Podman orchestration
- image resolution
- OCI digest recording
- mount policy
- stdout/stderr capture
- generated-file detection
- environment manifests
- image metadata
- tool version resolution
- container lifecycle
- disposable execution
- failure recovery

Introduce only enough tool environments to prove the architecture.

---

# M3 — Profiles and Transactional Host Management

Implement curated capability profiles.

Primary subjects:

- profile manifests
- dependency resolution
- RPM integration
- OCI profile components
- pre-change snapshot
- apply
- verification
- rollback
- profile state
- shell completion integration

Profiles:

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

Profiles may initially remain sparse.

The lightweight desktop direction is documented in `docs/DESKTOP.md`. Sway is
the intended tiling environment, but Leap 16 currently requires OBS packages
outside the official release repository. Repository identity, keys, exact
package versions, and rollback behavior must be modeled before the desktop is
added to a release image.

---

# M4 — Forensic Acquisition

Build conservative native acquisition workflows.

Primary subjects:

- device discovery
- mount detection
- software read-only handling
- disk metadata
- SMART
- NVMe metadata
- USB metadata
- partition inspection
- acquisition destinations
- streamed hashing
- imaging
- acquisition manifests
- chain-of-provenance records
- interruption handling

Conceptual commands:

    akilix acquire inspect
    akilix acquire protect
    akilix acquire image
    akilix acquire verify

Privilege boundaries become a major focus in M4.

---

# M5 — Forensic Analysis

Build forensic analysis environments.

Primary subjects:

- disk image handling
- filesystem analysis
- deleted-file analysis
- filesystem timelines
- registry/artifact parsing where applicable
- forensic container profiles
- artifact extraction
- provenance linking
- observations
- normalized timeline events

Analysis must preserve original evidence.

---

# M6 — Security Engineering Execution

Expand security-tool profiles after workbooks, provenance, and execution are mature.

Primary subjects:

- reconnaissance
- network analysis
- web assessment
- directory environments
- cloud environments
- wireless workflows
- blue-team tools
- incident-response tools

Scope awareness should integrate with Akilix-managed commands.

No milestone KPI should involve raw tool count.

---

# M7 — Findings, Timeline, and Reporting

Implement higher-level investigation objects.

Primary subjects:

- observations
- claims
- findings
- severity
- confidence
- evidence references
- normalized timelines
- report sections
- structured exports
- reproducible report generation

Findings should be traceable to artifacts and evidence.

---

# M8 — Local LLM Evidence Assistant

Introduce local-first analytical assistance.

Primary subjects:

- evidence broker
- searchable artifact index
- context construction
- LOCAL mode
- FACT/INFERENCE/HYPOTHESIS representation
- evidence citations
- timeline assistance
- finding drafting
- report assistance
- model provenance

The LLM is an assistant.

It is not an evidence generator.

---

# M9 — Controlled Remote LLM

Implement:

- REMOTE-REDACTED
- REMOTE-EXPLICIT

Primary subjects:

- redaction policy
- field classification
- transmission previews
- explicit consent
- remote-provider abstraction
- remote-request provenance
- audit records

Raw workbook evidence must never be silently transmitted.

---

# M10 — Workstation Experience

Polish Akilix as a professional daily workstation.

Primary subjects:

- Zsh environment
- workbook-aware prompt
- completions
- dynamic profile/tool completion
- TUI
- GUI
- evidence navigation
- findings navigation
- timeline UX
- report UX
- accessibility
- installation polish
- upgrade testing

The CLI/backend remains authoritative.

The passive, panel-based workbook workspace begins during M1. M10 adds the
interactive full-screen layer—focus, keyboard navigation, resizing, and richer
object views—without moving workbook business logic out of the backend.

---

# Long-term work

Potential later areas:

- dedicated forensic/live acquisition image
- analysis appliance
- hardware write-blocker integration
- remote team workbook synchronization
- cryptographic signing
- TPM-backed signing
- evidence sealing
- reproducible case export
- organization policy packs
- collaborative findings review
- isolated malware-analysis VMs
- automated openQA pipelines
- broader architecture support
