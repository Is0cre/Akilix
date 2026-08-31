# PenSUSE Playbooks

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

From the workbook TUI, press `p` to preview discovery. The operator supplies a
CIDR and an already-local image containing Nmap. PenSUSE resolves the immutable
digest, displays the complete plan, and requires the exact confirmation `RUN`.
It never pulls an image or starts discovery merely because the TUI was opened.

On completion, stdout and stderr are stored under `tool-output/`, scanner XML
is written beneath `artifacts/derived/<invocation-id>/`, and both the invocation
record and container policy manifest identify the scope decision and image
digest.
