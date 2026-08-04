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

    networks:
      include:
        - 10.40.0.0/16
      exclude:
        - 10.40.50.0/24

    domains:
      include:
        - example.org
        - "*.lab.example.org"
      exclude:
        - payments.example.org

    hosts:
      include:
        - vpn.example.org

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
