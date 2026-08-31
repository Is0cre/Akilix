# PenSUSE Workbooks

## Purpose

A workbook represents a specific authorized engagement, investigation, incident, case, or lab.

The workbook is the primary operational context of PenSUSE.

Examples:

    pensuse workbook create client-2026
    pensuse workbook open client-2026
    pensuse workbook overview client-2026

## Terminal navigation

`pensuse workbook open NAME` and `pensuse workbook overview NAME` present a
TUI-influenced workspace: workbook identity and lifecycle state, scope and
evidence counts, invocation health, explicit capture policy, canonical section
paths, and safe quick-jump commands. The view is deliberately passive and has
no terminal control sequences, so it remains useful when copied or redirected.
Use `overview --json` for future interactive TUI, GUI, or automation consumers;
those interfaces should reuse this model rather than reimplement discovery.

`pensuse workbook path NAME [SECTION]` prints exactly one absolute path for
shell composition. Sections are allowlisted and resolved only after opening a
valid workbook. For example:

    cd "$(pensuse workbook path client-2026 evidence)"
    cd "$(pensuse workbook path client-2026 original-evidence)"

Available sections are `root`, `artifacts`, `evidence`, `findings`, `logs`,
`notes`, `original-evidence`, `reports`, `timeline`, and `tool-output`.

---

# Identity

Each workbook has:

- immutable UUIDv7 identifier
- mutable human-readable name
- creation timestamp
- schema version

Renaming a workbook must never change its identity.

---

# Conceptual layout

    client-2026/
    ├── workbook.yaml
    ├── scope.yaml
    ├── logging.yaml
    ├── README.md
    ├── evidence/
    │   ├── original/
    │   ├── acquired/
    │   └── manifests/
    ├── artifacts/
    │   ├── imported/
    │   ├── derived/
    │   └── extracted/
    ├── captures/
    │   ├── network/
    │   ├── wireless/
    │   └── memory/
    ├── tool-output/
    ├── notes/
    ├── findings/
    ├── timeline/
    ├── reports/
    │   ├── drafts/
    │   └── exports/
    ├── logs/
    │   ├── command/
    │   ├── containers/
    │   └── audit/
    └── .pensuse/
        ├── index.sqlite
        ├── manifest.jsonl
        ├── state.json
        └── locks/

The filesystem remains canonical for evidence and primary provenance.

SQLite is an index and convenience layer.

---

# Workbook metadata

`workbook.yaml` should contain human-controlled workbook metadata.

Conceptual fields:

    schema: pensuse.workbook.v1
    id: 019c28b8-e4be-7ae5-a952-7ab97591a384
    name: client-2026
    created: 2026-08-04T18:00:00Z
    status: open
    description: Authorized assessment for client-2026

Additional fields should be introduced through schema versioning.

---

# Scope

`scope.yaml` describes authorized and excluded targets.

Potential scope classes:

- CIDRs
- IP addresses
- domains
- wildcard domains
- hosts
- applications/URLs
- exclusions
- notes
- timing restrictions

Scope is operator guidance and managed-execution protection.

It is not perfect containment.

---

# Logging policy

Workbook logging should be explicit and inspectable.

Example:

    Command logging         enabled
    Container logging       enabled
    Evidence hashing        enabled
    stdout capture          enabled
    stderr capture          enabled
    Packet metadata         disabled
    Full terminal capture   disabled

Changing logging policy should itself be auditable where appropriate.

---

# Object references

Workbooks should support explicit identifiers for:

- EV — Evidence
- AR — Artifact
- INV — Invocation
- OBS — Observation
- TL — Timeline Event
- CL — Claim
- FIND — Finding

Human-friendly display IDs may coexist with globally unique internal IDs.

---

# Passive opening

Opening a workbook must not:

- perform network scans
- resolve external resources unnecessarily
- run security tools
- upload case material
- open listeners

Workbook opening is a local state-selection operation.

---

# Workbook validation

`pensuse workbook validate NAME` is a local, read-only audit. It checks the
canonical workbook layout, parses scope and provenance records, and hashes every
original evidence object against its manifest. A mismatch fails validation and
does not rewrite evidence manifests. Use `pensuse evidence verify` when you
explicitly want persisted verification status updates.
