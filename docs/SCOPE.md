# PenSUSE Scope Model

## Purpose

Scope describes targets an operator is authorized to assess within a workbook.

PenSUSE uses scope information to reduce accidental mistakes in managed workflows.

Scope is not a perfect sandbox.

---

# Supported concepts

Initial scope model should support:

- IPv4 addresses
- IPv6 addresses
- CIDRs
- hostnames
- domains
- wildcard domains
- URLs/applications
- inclusions
- exclusions
- notes
- optional timing restrictions

---

# Evaluation

Structured results should include:

- ALLOW
- DENY
- UNKNOWN
- OVERRIDE

A specific exclusion should override a broad inclusion.

---

# Example

    version: 1

    include:
      - 10.40.0.0/16
      - example.org
      - "*.lab.example.org"
      - vpn.example.org
    exclude:
      - 10.40.50.0/24
      - payments.example.org

The initial v1 file stores canonical target rules in common include and exclude
lists. Target type is derived during local evaluation. A future typed scope
schema must use an explicit schema migration rather than silently changing this
on-disk representation.

---

# Overrides

Explicit operator overrides should be recorded in invocation provenance.

The record should preserve:

- original scope result
- override flag
- timestamp
- invocation reference

---

# Limits

PenSUSE cannot reliably infer every target a complex security tool may touch.

Documentation must not imply otherwise.
# Explainable decisions

`pensuse scope check WORKBOOK TARGET --json` returns the normalized decision and,
when applicable, the canonical include or exclude rule that matched it:

```json
{"target":"10.0.5.2","result":"DENY","rule":"10.0.5.0/24"}
```

Exclusions are evaluated first, so a denied exclusion remains authoritative even
when a broader include also matches. The decision record is local and does not
imply a network probe or a security sandbox.
