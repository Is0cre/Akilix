#!/bin/sh
# Statically validate the live ISO network baseline without changing host state.
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

python3 - <<'PY'
import fnmatch
import xml.etree.ElementTree as ET
from pathlib import Path

image = Path("image/kiwi-iso")
config = image / "config.xml"
preset_dir = image / "root/etc/systemd/system-preset"

packages = {
    node.get("name")
    for node in ET.parse(config).findall(".//package")
    if node.get("name")
}
required = {
    "NetworkManager": "NetworkManager",
    "nmtui (NetworkManager-tui)": "NetworkManager-tui",
    "iputils": "iputils",
    "nmap": "nmap",
}
missing = [label for label, package in required.items() if package not in packages]
if missing:
    raise SystemExit("KIWI network baseline lacks: " + ", ".join(sorted(missing)))


def effective_preset(unit):
    """Return systemd's first matching preset rule in lexical file order."""
    for path in sorted(preset_dir.glob("*.preset")):
        for number, raw in enumerate(path.read_text().splitlines(), 1):
            line = raw.strip()
            if not line or line.startswith("#"):
                continue
            fields = line.split()
            if len(fields) < 2 or fields[0] not in {"enable", "disable", "ignore"}:
                raise SystemExit(f"invalid preset rule at {path}:{number}: {raw}")
            if fnmatch.fnmatchcase(unit, fields[1]):
                return fields[0], path, number
    return None


expected = {
    "NetworkManager.service": "enable",
    "wicked.service": "disable",
}
for unit, wanted in expected.items():
    result = effective_preset(unit)
    if result is None:
        raise SystemExit(f"no image preset selects {unit}")
    action, path, number = result
    if action != wanted:
        raise SystemExit(
            f"effective preset for {unit} is {action}, expected {wanted} "
            f"({path}:{number})"
        )

print("network image check: packages and service presets valid")
PY
