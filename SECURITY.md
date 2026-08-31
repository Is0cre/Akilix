# Akilix Security Policy

## Status

Akilix is under active development and is not yet production-ready.

Do not rely on development builds as the sole protection mechanism for forensic evidence or production security work.

---

# Security principles

Akilix follows these baseline principles:

- least privilege
- explicit operator action
- local-first case data
- no telemetry by default
- no hidden monitoring
- no automatic scans or attacks
- no automatically opened listeners
- reproducible execution environments
- immutable identification of managed OCI images
- separation of original evidence and derived material
- transparent logging
- conservative forensic acquisition
- clear distinction between safety guards and security boundaries

---

# Evidence safety

Original evidence must not be modified during analysis.

Evidence integrity must be supported by cryptographic hashes.

Btrfs snapshots do not substitute for cryptographic evidence verification.

Software write protection does not substitute for hardware forensic write blockers.

---

# Privilege model

The main `akilix` process should run unprivileged for ordinary operations.

Privileged actions must be explicit and narrowly scoped.

Akilix must not solve convenience problems by making the primary process permanently root.

---

# Container security

Rootless Podman is preferred.

Managed tool containers should avoid:

- `--privileged`
- host PID namespace
- unrestricted host filesystem mounts
- arbitrary device passthrough
- unnecessary host networking

Original evidence mounts should be read-only during analysis where technically practical.

---

# Network behavior

Akilix itself must not silently:

- scan
- attack
- listen
- upload
- phone home
- activate telemetry

External service integration must be explicit.

---

# LLM safety

LLM output is analysis, not evidence.

Akilix must distinguish:

- FACT
- INFERENCE
- HYPOTHESIS

Remote LLM use must not silently transmit case material.

`REMOTE-EXPLICIT` means operator-approved transmission for the selected operation.

---

# Reporting vulnerabilities

Until a formal security contact is established, security issues should be reported privately to the project maintainer rather than disclosed in public issue trackers when disclosure would put users at risk.

A dedicated reporting channel and coordinated disclosure policy should be established before the first public security-sensitive release.

---

# Threat model

See:

- `docs/PRIVILEGE.md`
- `docs/EVIDENCE.md`
- `docs/CONTAINERS.md`
- `docs/INVARIANTS.md`

for relevant architectural security requirements.
