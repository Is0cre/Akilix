#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

python3 - <<'PY'
import re
import xml.etree.ElementTree as ET
from pathlib import Path

image = Path("image/kiwi-iso")
zsh = (image / "root/etc/zsh.zshrc.local").read_text().splitlines()
active = [line.strip() for line in zsh if line.strip() and not line.lstrip().startswith("#")]
for line in active:
    if re.search(r"(^|\s)(preexec|command_not_found_handler)(\s|=|\()", line):
        raise SystemExit(f"operator shell installs command interception: {line}")
    if re.match(r"alias\s+(nmap|tcpdump|podman|zypper|rpm|ssh|curl)=", line):
        raise SystemExit(f"operator tool is wrapped by a system alias: {line}")
    if re.search(r"exec\s+akilix(\s|$)", line):
        raise SystemExit(f"operator shell is forced into Akilix: {line}")

sway = (image / "root/etc/sway/config").read_text()
for marker in ("set $term foot", "exec $term --title \"Akilix Workspace\"", "bindsym $mod+Return exec $term", "bindsym Ctrl+Alt+t exec $term"):
    if marker not in sway:
        raise SystemExit(f"desktop lacks unrestricted shell access: {marker}")

packages = {node.get("name") for node in ET.parse(image / "config.xml").findall(".//package")}
direct = {"nmap", "tcpdump", "curl", "git", "vim", "nano", "podman", "iproute2", "iputils"}
if not direct <= packages:
    raise SystemExit(f"base image lacks direct operator tools: {sorted(direct - packages)}")
PY

echo "operator freedom check: direct shell and native tools remain available"
