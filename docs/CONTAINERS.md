# Akilix Container Model

## Runtime

Podman is the default OCI runtime.

Rootless execution is preferred.

Check the local runtime without contacting a registry:

    akilix container doctor

The check requires Podman to report rootless operation and verifies that the
operator's user namespace can be entered. It does not pull or build an image.

---

# Purpose

Containers isolate fast-moving or conflicting security tooling from the conservative host.

They are not the location where workbook evidence permanently lives.

Containers are disposable execution environments.

---

# Tool families

Initial conceptual environments include:

- recon
- network
- web
- directory
- cloud
- reverse
- malware
- forensics

Akilix does not require one container per tool.

Coherent family images are preferred where practical.

---

# Reproducibility

Managed execution records:

- image reference
- immutable OCI digest
- tool name
- tool version where determinable
- argument vector
- start/end timestamps
- exit status
- stdout/stderr
- generated files
- workbook ID

Mutable tags such as `latest` are insufficient provenance.

---

# Mount policy

Typical analysis mounts should resemble:

    /workbook/evidence/original   read-only
    /workbook/artifacts           writable
    /workbook/tool-output         writable
    /workbook/tmp                 writable

Mount only what the operation requires.

---

# Default restrictions

Generated container specs currently default to:

- `--pull=never` with an immutable image digest
- `--network=none`
- private PID, IPC, and UTS namespaces
- `--userns=keep-id`
- read-only root filesystem
- `no-new-privileges`
- all Linux capabilities dropped

Original-evidence mounts must be explicitly read-only. Writable mounts are
reserved for derived evidence, tool output, and temporary workspaces.

Original evidence is not mounted by default. The operator must request the
fixed policy explicitly:

    akilix container run CASE IMAGE --mount-originals -- TOOL /workbook/evidence/original/item

Akilix validates that the canonical source directory has not been replaced or
redirected by a symlink and always mounts it read-only at
`/workbook/evidence/original`. Arbitrary host paths are not accepted.

Writable output is also opt-in. `--mount-output` creates a fresh
`artifacts/derived/<invocation-id>/` directory and mounts it at
`/workbook/output`. Files remaining there after successful or failed execution
are attributed to that invocation. Symlinks are rejected rather than recorded
as valid generated artifacts.

Podman options, including every mount, are emitted before the immutable image
reference. This ordering is security-relevant: options after the image would
instead become arguments to the containerized command.

Do not automatically use:

- `--privileged`
- host PID namespace
- arbitrary `/dev` passthrough
- unrestricted host filesystem
- unrestricted capabilities
- host network mode

Privilege expansion must be deliberate and visible.

---

# Networking

Some security tools require network access.

Network mode should be selected according to the tool/workflow rather than always granting maximum access.

Scope checks remain an operator safety layer, not a perfect network sandbox.

---

# Image supply chain

Long-term container infrastructure should support:

- deterministic definitions
- package inventories
- SBOM generation
- OCI digests
- signatures
- provenance
- local mirrors/registries
- offline import/export

Akilix should remain operable with self-hosted OCI infrastructure.

---

# Evidence boundary

Generic privileged containers are not the architecture for raw forensic acquisition.

Physical acquisition belongs to the dedicated acquisition subsystem.
