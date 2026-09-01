# Akilix Workbooks

## Purpose

A workbook represents a specific authorized engagement, investigation, incident, case, or lab.

The workbook is the primary operational context of Akilix.

Examples:

    akilix workbook create client-2026
    akilix workbook open client-2026
    akilix workbook overview client-2026

## Terminal navigation

`akilix workbook open NAME` and `akilix workbook overview NAME` present a
TUI-influenced workspace: workbook identity and lifecycle state, scope and
evidence counts, invocation health, explicit capture policy, canonical section
paths, and safe quick-jump commands. The view is deliberately passive and has
no terminal control sequences, so it remains useful when copied or redirected.
Use `overview --json` for future interactive TUI, GUI, or automation consumers;
those interfaces should reuse this model rather than reimplement discovery.

`akilix workbook path NAME [SECTION]` prints exactly one absolute path for
shell composition. Sections are allowlisted and resolved only after opening a
valid workbook. For example:

    cd "$(akilix workbook path client-2026 evidence)"
    cd "$(akilix workbook path client-2026 original-evidence)"

Available sections are `root`, `artifacts`, `evidence`, `findings`, `hardware`,
`logs`, `notes`, `original-evidence`, `reports`, `timeline`, and `tool-output`.

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
    ├── journal.jsonl
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
    ├── hardware/
    │   ├── inspections/
    │   ├── identifications/
    │   └── protections/
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
    └── .akilix/
        ├── index.sqlite
        ├── manifest.jsonl
        ├── state.json
        └── locks/

The filesystem remains canonical for evidence and primary provenance.

`journal.jsonl` is the unified operational projection for rapid human review.
It uses flat, bounded `akilix.journal.v1` JSONL records with millisecond UTC
timestamps, semantic event/module tokens, payloads, and provenance IDs.
Canonical evidence, invocation, hardware, and scope records remain authoritative;
the journal does not replace them. Writers use process-local serialization,
an inter-process file lock, append mode, and `fsync`.

Managed native commands appear as `ENGINEERING` events, while container runs
remain `OCI`. This records lifecycle metadata for explicit `akilix run`
invocations; it does not silently capture arbitrary shell commands or enable
terminal recording. Evidence imports journal an explicit request followed by
`EVIDENCE_IMPORTED` or `EVIDENCE_IMPORT_FAILED`, and each independent hash check
adds an `EVIDENCE_VERIFIED` event carrying its match or mismatch result. The
canonical evidence manifest and invocation manifest remain authoritative.
Explicit close, reopen, and rename operations likewise produce request/result
pairs. Passive open, list, status, overview, and validation reads do not append
events or otherwise mutate the workbook.

## Live activity window

Inside the interactive workbook, `[L]` opens a native Go viewport over the
unified journal. It does not launch `tail`. The reader processes bounded chunks,
does not retain a descriptor between polls, keeps at most 500 rendered records,
and returns to the action menu with Escape or `q`.

`akilix workbook follow NAME` renders new canonical invocation records as they
are completed. `akilix workbook follow NAME --once` prints the current
snapshot and exits, which is useful for scripts and diagnostics. In the
workbook TUI, the `L` action explicitly opens the follower in a separate Foot window so
Sway can tile it beside the operator workspace.

For each confirmed TUI playbook run under Sway, Akilix assigns a dedicated
Foot window to a workspace named for the tool and the final four invocation-ID
characters, such as `naabu-4a2c`. That window filters lifecycle events to the
invocation and follows its captured stdout and stderr while the workbook-wide
activity window continues aggregating every managed source. Outside Sway, the
same backend execution has no desktop side effect.

Following is a local, read-only operation. It does not execute tools, inspect
the network, capture the terminal, or change the workbook. The versioned
`.akilix/activity.jsonl` journal emits `STARTED`, `COMPLETED`, and `FAILED`
events for managed native and container executions. It is an operator activity
stream; the final invocation manifest remains the canonical provenance record.

SQLite is an index and convenience layer.

---

# Workbook metadata

`workbook.yaml` should contain human-controlled workbook metadata.

Conceptual fields:

    schema: akilix.workbook.v1
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

`akilix workbook validate NAME` is a local, read-only audit. It checks the
canonical workbook layout, parses scope and provenance records, and hashes every
original evidence object against its manifest. A mismatch fails validation and
does not rewrite evidence manifests. Use `akilix evidence verify` when you
explicitly want persisted verification status updates.
