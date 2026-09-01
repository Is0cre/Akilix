# Akilix Playbooks

Playbooks are typed Go plans shared by the CLI, TUI, and future interfaces.
They do not contain shell strings and planning never performs the described
activity. Execution requires a separate, explicit operator action.

## Local network discovery

`local-network-discovery` accepts one canonical CIDR that must be explicitly
allowed by the selected open workbook. It translates IP and CIDR exclusions
contained by that network into scanner exclusions. Original evidence is not
mounted.

The initial command vector uses conservative host discovery (`nmap -sn -n`),
not vulnerability scanning or exploitation. Opening a workbook or TUI never
runs it. A preview must show scope reasoning, exclusions, immutable image,
network mode, mounts, arguments, and output path before confirmation.

Semantic attention values (`INFO`, `SAFE`, `WARN`, and `BLOCK`) are data, not
hard-coded ANSI colors. The TUI maps them to the current accessible palette so
meaning remains available through labels as well as color.

From the workbook TUI, press `n`, supply an explicitly allowed CIDR or IP, and
press Enter to start the labeled discovery action. Interactive discovery uses
the Nmap binary shipped in the base image. It writes XML to standard output so
the managed native invocation engine captures it together with stderr, exact
argv, timestamps, exit state, workbook identity, and scope decision. It does
not invoke a shell or perform DNS resolution. Opening the workbook never starts
discovery.

On completion, stdout and stderr are stored under `tool-output/`; stdout is the
scanner XML artifact. The invocation record identifies its scope decision.
The digest-pinned OCI Nmap plan remains available as planning logic for a
future profile-controlled execution option.

Akilix then streams the completed XML artifact through a bounded native parser
without modifying it. Every reported IPv4 or IPv6 address is evaluated again
through the canonical scope engine. Allowed addresses become
`HOST_DISCOVERED` journal observations linked to the invocation; unexpected
addresses become `HOST_DROPPED_OUT_OF_SCOPE` alerts and are never silently
promoted as actionable targets. MAC addresses and hosts marked down are not
converted into discovery observations.

## Local port discovery

`local-port-discovery` uses the native Nmap included in the base image. Press
`p` in the workbook TUI and explicitly submit an allowed CIDR or IP. The
managed argv selects unprivileged TCP CONNECT scanning, the top 100 ports, a
maximum rate of 100 probes per second, one retry, a 30-second per-host timeout,
no DNS resolution, and XML output captured as invocation stdout.
The workbook TUI applies a 30-minute overall deadline to both native discovery
actions. Pressing Escape while a scan is running cancels its context; the
managed invocation remains recorded as failed/interrupted with captured output
instead of becoming an invisible orphan process.

The completed XML is parsed through the same bounded, symlink-safe ingester as
host discovery. Only ports explicitly marked open become `PORT_FOUND` events;
each carries the address, endpoint, protocol, optional service label, scope
decision, journal provenance, and originating invocation ID. Every result is
rechecked against current workbook scope before presentation.

The existing Naabu plan remains available for the optional digest-pinned
`localhost/local-naabu` profile execution path; it is no longer required for
the first-boot TUI workflow.

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

After a successful managed run, Akilix parses that attributed JSONL artifact
without scraping terminal output. In-scope sockets become `PORT_FOUND` journal
events linked to the invocation. Results that the canonical scope engine does
not allow become `PORT_DROPPED_OUT_OF_SCOPE` events and are not presented as
actionable discoveries. The original tool artifact remains unchanged.
