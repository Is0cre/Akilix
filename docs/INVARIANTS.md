# Akilix Invariants

These rules define behavior that Akilix must preserve across milestones.

They should eventually become automated tests wherever technically feasible.

---

## PS-INV-001 — Workbook opening is passive

Opening or inspecting a workbook MUST NOT automatically:

- scan a network
- resolve unrelated remote resources
- upload data
- invoke security tools
- open listeners
- execute engagement activity

---

## PS-INV-002 — Original evidence is immutable during analysis

Analysis operations MUST NOT modify original evidence.

Original evidence must be clearly distinguished from writable derived material.

---

## PS-INV-003 — Evidence imports never silently overwrite

Importing evidence MUST NOT silently replace an existing evidence object or original evidence file.

---

## PS-INV-004 — Cryptographic evidence verification

Evidence objects MUST support SHA-256 verification.

A verification result must distinguish:

- match
- mismatch
- incomplete/unavailable verification

---

## PS-INV-005 — Provenance is explicit

Managed security operations MUST NOT depend solely upon shell history for reconstruction.

---

## PS-INV-006 — Container identity is immutable

Managed container invocation records MUST resolve and store an immutable OCI image digest.

Mutable tags alone are insufficient provenance.

---

## PS-INV-007 — Scope is not represented as containment

Scope checking MUST NOT be described as providing perfect sandboxing or technical enforcement of engagement authorization.

---

## PS-INV-008 — Scope overrides are visible

An explicit override of a Akilix scope warning MUST be represented in invocation provenance.

---

## PS-INV-009 — No hidden telemetry

Akilix MUST NOT transmit telemetry by default.

---

## PS-INV-010 — No hidden case upload

Workbook content MUST NOT be transmitted remotely unless a feature explicitly selected by the operator requires it.

---

## PS-INV-011 — Transparent logging

The operator MUST be able to inspect which logging features are active.

---

## PS-INV-012 — LLM output is not evidence

LLM-generated statements MUST NOT automatically become evidence objects.

---

## PS-INV-013 — LLM epistemic classification

LLM-assisted analysis MUST support distinction between:

- FACT
- INFERENCE
- HYPOTHESIS

---

## PS-INV-014 — Evidence-backed factual claims

Where an LLM presents a factual claim derived from workbook material, Akilix SHOULD preserve source references sufficient to inspect the supporting material.

---

## PS-INV-015 — Explicit remote evidence transfer

REMOTE-EXPLICIT LLM operation MUST require explicit operator approval for the selected transmission.

---

## PS-INV-016 — Host rollback does not roll back evidence

Operating-system rollback MUST NOT silently revert workbook evidence.

System snapshots and workbook data require separate lifecycle handling.

---

## PS-INV-017 — Snapshots are not evidence integrity

Akilix documentation and UI MUST NOT represent Btrfs snapshots as a substitute for evidence hashing.

---

## PS-INV-018 — Software write protection is accurately described

Software read-only handling MUST NOT be represented as equivalent to a hardware forensic write blocker.

---

## PS-INV-019 — Acquisition is separate from generic tool execution

Raw physical-device acquisition MUST NOT depend on generic unrestricted privileged security containers.

---

## PS-INV-020 — Least privilege

Ordinary Akilix operations SHOULD execute without root privileges.

Privileged operations MUST be explicit.

---

## PS-INV-021 — No automatic attacks

Akilix MUST NOT automatically perform attacks because a workbook, profile, or application was opened.

---

## PS-INV-022 — No automatic listeners

Installing or enabling a Akilix security profile MUST NOT silently expose network listeners.

---

## PS-INV-023 — Tool count is not a requirement

No milestone may use the number of bundled security tools as a success criterion.

---

## PS-INV-024 — Canonical evidence survives index loss

Loss or corruption of a local search/index database MUST NOT make original evidence or its primary provenance unintelligible.

---

## PS-INV-025 — Failed operations are distinguishable

Interrupted or failed evidence/provenance operations MUST NOT leave records that appear indistinguishable from successfully completed operations.

---

## PS-INV-026 — Explicit command arguments

Managed tool execution SHOULD use structured argument vectors rather than constructing shell command strings.

---

## PS-INV-027 — Workbooks have immutable identity

Renaming the human-readable workbook name MUST NOT change its immutable workbook identity.

---

## PS-INV-028 — Generated artifacts are attributable

Artifacts produced by managed executions SHOULD be attributable to their originating invocation.

---

## PS-INV-029 — Tool versions are attributable

Managed executions SHOULD record tool versions where they can be reliably determined.

---

## PS-INV-030 — Host modifications are recoverable

Major Akilix-managed host/profile changes SHOULD use:

    SNAPSHOT
    APPLY
    VERIFY

where technically practical.

---

## PS-INV-031 — No mandatory external service

Core Akilix functionality MUST NOT require an external SaaS service, hosted control plane, telemetry endpoint, or remote LLM.

---

## PS-INV-032 — General-purpose operator access

Akilix MUST remain a usable openSUSE workstation outside a workbook. The
operator MUST be able to open an ordinary shell and invoke installed native
tools directly without creating or selecting a workbook.

Akilix MUST NOT intercept, silently rewrite, scope-block, or automatically
journal arbitrary shell commands. Direct execution is explicitly unmanaged:
it does not receive workbook scope assistance, artifact attribution, or
canonical invocation provenance unless the operator chooses an Akilix-managed
execution path.

---

# Invariant changes

Removing or weakening an invariant is an architectural change.

Such changes must not occur incidentally during feature implementation.
