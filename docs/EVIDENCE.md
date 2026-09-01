# Akilix Evidence Model

## Principle

Original evidence is preserved.

Analysis operates on evidence without modifying the original source.

---

# Classes

Akilix should distinguish:

## Original evidence

Material imported or acquired as the authoritative source.

Examples:

- disk image
- memory dump
- packet capture
- firmware image
- exported log archive

## Acquired evidence

Evidence created by Akilix acquisition workflows.

## Derived artifact

Material produced from evidence.

Examples:

- filesystem timeline
- extracted executable
- parsed registry data
- converted packet data
- strings output

## Tool output

stdout, stderr, reports, parser output, and generated files associated with an invocation.

---

# Hashes

SHA-256 is mandatory for evidence objects.

Stored metadata should include:

- algorithm
- digest
- byte size
- calculation start time
- calculation completion time
- verification status

Additional algorithms may be included for interoperability.

---

# Provenance

Evidence metadata should answer:

- where did this come from?
- when was it acquired/imported?
- by which operation?
- which tool/version was involved?
- which source device/file was used?
- what hashes were calculated?
- what artifacts were derived from it?

---

# Sidecar metadata

Evidence should have machine-readable sidecar metadata rather than relying only on a central database.

Conceptual example:

    {
      "schema": "akilix.evidence.v1",
      "evidence_id": "019c...",
      "classification": "original",
      "filename": "disk01.E01",
      "size": 512110190592,
      "hashes": {
        "sha256": "..."
      }
    }

---

# Filesystem safety

Evidence handling code must consider:

- symlinks
- hard links
- path traversal
- pre-existing destinations
- interrupted writes
- disk full
- mount boundaries
- ownership
- permissions

Evidence imports must never silently overwrite existing originals.

---

# Read-only analysis

Where technically practical:

- original evidence mounts are read-only
- containers receive originals read-only
- writable outputs use separate directories

Read-only mount policy is a safety/security control but does not replace hashing.

Native block-device acquisition writes operator-named images beneath
`evidence/acquired/`. Akilix publishes an image only after the full expected
device size has streamed successfully, the destination has been fsynced, and a
SHA-256 digest has been calculated inline. REQUESTED, COMPLETED, and FAILED
records under `hardware/acquisitions/` make interrupted operations visibly
different from completed acquisitions. Kernel software read-only state is
revalidated immediately before imaging but is not represented as equivalent to
a hardware forensic write blocker.

---

# Device acquisition

Software read-only controls should be attempted and verified where technically possible.

Akilix must clearly distinguish software protection from hardware forensic write blockers.

The workbook TUI `[H] Hardware inventory` view performs the same passive,
structured `lsblk` inspection as `akilix acquire inspect`. It labels the host
system disk separately from acquisition candidates and shows kernel read-only,
mount, removable-media, transport, and local trust state. Opening or refreshing
this view does not open a block-device node, mount media, record provenance, or
change device state. Identification, protection, and imaging remain explicit
acquisition actions.

---

# Recovery

Loss of `index.sqlite` must not render evidence unintelligible.

Primary evidence metadata must remain available in canonical workbook files.
