# Akilix Playbooks

Playbooks are typed Go plans shared by the CLI, TUI, and future interfaces.
They do not contain shell strings and planning never performs the described
activity. Execution requires a separate, explicit operator action.

## Local network discovery

`local-network-discovery` accepts one canonical CIDR that must be explicitly
allowed by the selected open workbook. It translates IP and CIDR exclusions
contained by that network into scanner exclusions, uses a digest-pinned image,
requests bridge networking, and writes XML to the invocation-specific
`/workbook/output` mount. Original evidence is not mounted.

The initial command vector uses conservative host discovery (`nmap -sn -n`),
not vulnerability scanning or exploitation. Opening a workbook or TUI never
runs it. A preview must show scope reasoning, exclusions, immutable image,
network mode, mounts, arguments, and output path before confirmation.

Semantic attention values (`INFO`, `SAFE`, `WARN`, and `BLOCK`) are data, not
hard-coded ANSI colors. The TUI maps them to the current accessible palette so
meaning remains available through labels as well as color.

From the workbook TUI, press `n` to preview discovery. The operator supplies a
CIDR and an already-local image containing Nmap. Akilix resolves the immutable
digest, displays the complete plan, and requires the exact confirmation `RUN`.
It never pulls an image or starts discovery merely because the TUI was opened.

On completion, stdout and stderr are stored under `tool-output/`, scanner XML
is written beneath `artifacts/derived/<invocation-id>/`, and both the invocation
record and container policy manifest identify the scope decision and image
digest.

## Local port discovery

`local-port-discovery` is a separate Naabu-backed playbook; it does not replace
Nmap. Press `p` in the workbook TUI, provide an explicitly allowed CIDR and an
already-local Naabu image, inspect the complete plan, and type the exact word
`RUN` to begin.

The Naabu plan supplies flags to the image's `naabu` entrypoint, matching the
official ProjectDiscovery image contract. A generic image without that
entrypoint is not compatible with this playbook.

The initial policy uses an unprivileged TCP CONNECT scan of the top 100 ports,
limits the rate to 100 connections per second, uses one retry and a one-second
timeout, and carries contained workbook exclusions into `-exclude-hosts`. It
disables Naabu's update check and does not select passive discovery, cloud
upload, SYN scanning, host networking, or added capabilities. Results are
written as JSONL to the invocation-specific
`/workbook/output/ports.jsonl`; stdout, stderr, scope reasoning, exact argument
vector, and immutable image digest remain attributable through the normal
container invocation record.
