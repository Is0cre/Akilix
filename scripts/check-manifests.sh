#!/bin/sh
# Validate repository-owned machine-readable manifests without changing host state.
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

if grep -R "oakilix" --exclude-dir=.git --exclude-dir=build . >/dev/null 2>&1; then
	echo "corrupted openSUSE origin found after identity substitution" >&2
	exit 1
fi

for schema in schemas/*.json; do
	python3 -m json.tool "$schema" >/dev/null
done

# Use the built CLI when available so this check exercises the shipped binary.
# Fall back to go run for direct developer use before `make build`.
cli=./akilix
if [ ! -x "$cli" ]; then
	cli="go run ./cmd/akilix"
fi

repositories_json=$($cli repository list --json)
REPOSITORIES_JSON="$repositories_json" python3 - <<'PY'
import json
import os

items = json.loads(os.environ["REPOSITORIES_JSON"])
if not isinstance(items, list) or not items:
    raise SystemExit("repository list must be a non-empty JSON array")
for item in items:
    if not item.get("id") or not item.get("key_fingerprint"):
        raise SystemExit("repository entry lacks identity or signing key")
    if item.get("image_enabled") and item.get("status") != "approved":
        raise SystemExit("image-enabled repository is not approved")
PY

python3 - <<'PY'
import json
from pathlib import Path

lock = json.loads(Path("repositories/desktop-sway-lock.json").read_text())
expected = {"sway", "swaybar", "swayidle", "swaylock", "foot", "fuzzel", "mako", "wl-clipboard", "grim", "slurp", "greetd", "greetd-branding-upstream", "tuigreet"}
packages = lock.get("packages", [])
names = {item.get("name") for item in packages}
if lock.get("schema") != "akilix.package-lock.v1" or names != expected:
    raise SystemExit("desktop package lock is incomplete")
for item in packages:
    location = item.get("location", "")
    if len(item.get("sha512", "")) != 128 or not location.startswith(("x86_64/", "noarch/")):
        raise SystemExit("desktop package lock contains invalid RPM identity")
PY

python3 - <<'PY'
import json
from pathlib import Path

lock = json.loads(Path("repositories/desktop-fonts-lock.json").read_text())
expected = {"google-noto-sans-fonts", "google-noto-sans-mono-fonts", "google-noto-sans-symbols2-fonts", "google-noto-coloremoji-fonts"}
packages = lock.get("packages", [])
if lock.get("schema") != "akilix.package-lock.v1" or lock.get("repository_id") != "m17n-fonts-leap-16":
    raise SystemExit("desktop font lock has the wrong source identity")
if {item.get("name") for item in packages} != expected:
    raise SystemExit("desktop font lock is incomplete")
for item in packages:
    if len(item.get("sha512", "")) != 128 or not item.get("location", "").startswith("noarch/"):
        raise SystemExit("desktop font lock contains invalid RPM identity")
PY

python3 - <<'PY'
import json
from pathlib import Path

lock = json.loads(Path("repositories/boot-plymouth-lock.json").read_text())
expected = {"plymouth", "plymouth-branding-upstream", "plymouth-dracut", "plymouth-plugin-script"}
packages = lock.get("packages", [])
if lock.get("schema") != "akilix.package-lock.v1":
    raise SystemExit("Plymouth package lock has the wrong schema")
if lock.get("repository_id") != "base-system-leap-16":
    raise SystemExit("Plymouth package lock has the wrong repository")
if {item.get("name") for item in packages} != expected:
    raise SystemExit("Plymouth package lock is incomplete")
for item in packages:
    if len(item.get("sha512", "")) != 128 or not item.get("location", "").startswith(("x86_64/", "noarch/")):
        raise SystemExit("Plymouth package lock contains invalid RPM identity")
PY

python3 - <<'PY'
import tomllib
import xml.etree.ElementTree as ET
from pathlib import Path

image = Path("image/kiwi-iso")
config = tomllib.loads((image / "root/etc/greetd/config.toml").read_text())
session = config.get("default_session", {})
if config.get("terminal", {}).get("vt") != 1:
    raise SystemExit("greetd must own tty1")
if session.get("user") != "greeter" or "--cmd sway" not in session.get("command", ""):
    raise SystemExit("greetd does not launch the audited Sway session")
if "--remember" in session.get("command", ""):
    raise SystemExit("greetd must not remember operator identity by default")
command = session.get("command", "")
for setting in (
    "--issue",
    "--theme",
    "--sessions /usr/share/wayland-sessions",
    "--power-shutdown '/usr/bin/systemctl poweroff'",
    "--power-reboot '/usr/bin/systemctl reboot'",
):
    if setting not in command:
        raise SystemExit(f"greetd lacks required Akilix setting: {setting}")
issue = (image / "root/etc/issue").read_text()
for marker in ("AKILIX", "Security work with provenance.", "Live development image"):
    if marker not in issue:
        raise SystemExit(f"greetd issue banner lacks: {marker}")

packages = {node.get("name") for node in ET.parse(image / "config.xml").findall(".//package")}
required = {"greetd", "greetd-branding-upstream", "tuigreet", "sway", "swaybar"}
if not required <= packages:
    raise SystemExit("KIWI image lacks the complete greetd/Sway session")

links = {
    image / "root/etc/systemd/system/default.target": "/usr/lib/systemd/system/graphical.target",
    image / "root/etc/systemd/system/display-manager.service": "/usr/lib/systemd/system/greetd.service",
}
for path, target in links.items():
    if not path.is_symlink() or str(path.readlink()) != target:
        raise SystemExit(f"invalid systemd image link: {path}")
if (image / "root/etc/profile.d/50-akilix-sway.sh").exists():
    raise SystemExit("legacy shell-profile Sway autostart is still present")

tree = ET.parse(image / "config.xml")
operator = tree.find(".//user[@name='akilix']")
if operator is None or operator.get("shell") != "/bin/zsh":
    raise SystemExit("live operator does not use Zsh by default")
repositories = {node.get("alias"): node.find("source").get("path") for node in tree.findall("repository")}
if repositories.get("Akilix-Base-System-Leap-16") != "https://download.opensuse.org/repositories/Base:/System/16.0/":
    raise SystemExit("KIWI image lacks the audited Base:System source")
if repositories.get("Akilix-M17N-Fonts-Leap-16") != "https://download.opensuse.org/repositories/M17N:/fonts/16.0/":
    raise SystemExit("KIWI image lacks the audited M17N:fonts source")
plymouth = {"plymouth", "plymouth-branding-upstream", "plymouth-plugin-script", "plymouth-dracut"}
if not plymouth <= packages:
    raise SystemExit("KIWI image lacks the complete Plymouth/initrd package set")
preferences = tree.find("preferences")
image_type = preferences.find("type")
if preferences.findtext("bootsplash-theme") != "akilix":
    raise SystemExit("KIWI does not select the Akilix Plymouth theme")
if "splash" not in image_type.get("kernelcmdline", "").split():
    raise SystemExit("KIWI kernel command line does not activate Plymouth")

fonts = {"google-noto-sans-fonts", "google-noto-sans-mono-fonts", "google-noto-sans-symbols2-fonts", "google-noto-coloremoji-fonts"}
if not fonts <= packages:
    raise SystemExit("KIWI image lacks the selected desktop font set")
foot = (image / "root/etc/xdg/foot/foot.ini").read_text()
if "font=Noto Sans Mono:size=11,Noto Sans Symbols 2:size=11,Noto Color Emoji:size=11" not in foot:
    raise SystemExit("Foot does not use the installed Noto font stack")
sway = (image / "root/etc/sway/config").read_text()
for setting in ("status_command /usr/bin/akilix bar stream", "focused_workspace #657a3e"):
    if setting not in sway:
        raise SystemExit(f"Sway native command strip lacks {setting}")
if "exec waybar" in sway or "waybar" in packages:
    raise SystemExit("legacy Waybar integration is still active")

zshrc = (image / "root/etc/zsh.zshrc.local").read_text()
for setting in ("HISTSIZE=100000", "SAVEHIST=75000", "EXTENDED_HISTORY", "HIST_IGNORE_SPACE", "SHARE_HISTORY"):
    if setting not in zshrc:
        raise SystemExit(f"Zsh baseline lacks {setting}")
PY

zsh -n image/kiwi-iso/root/etc/zsh.zshrc.local

profile_dir=${AKILIX_PROFILE_DIR:-profiles}
profiles_json=$(AKILIX_PROFILE_DIR="$profile_dir" $cli profile list --json)
PROFILE_JSON="$profiles_json" python3 - <<'PY'
import json
import os
import sys

try:
    profiles = json.loads(os.environ["PROFILE_JSON"])
except (KeyError, json.JSONDecodeError) as exc:
    print(f"invalid profile list JSON: {exc}", file=sys.stderr)
    raise SystemExit(1)

if not isinstance(profiles, list):
    print("profile list must be a JSON array", file=sys.stderr)
    raise SystemExit(1)

for item in profiles:
    if not isinstance(item, dict) or not item.get("id"):
        print("profile list contains an invalid entry", file=sys.stderr)
        raise SystemExit(1)
PY

# Exercise both read paths for every manifest; this catches malformed plans and
# keeps the audit aligned with the operator-facing command surface.
profile_ids=$(printf '%s\n' "$profiles_json" | python3 -c 'import json,sys; print("\n".join(item["id"] for item in json.load(sys.stdin)))')
for profile_id in $profile_ids; do
	show_json=$(AKILIX_PROFILE_DIR="$profile_dir" $cli profile show "$profile_id" --json)
	plan_json=$(AKILIX_PROFILE_DIR="$profile_dir" $cli profile plan "$profile_id" --json)
	SHOW_JSON="$show_json" PLAN_JSON="$plan_json" python3 - <<'PY'
import json
import os
import sys

show = json.loads(os.environ["SHOW_JSON"])
plan = json.loads(os.environ["PLAN_JSON"])
if not isinstance(show, dict) or not show.get("id"):
    print("profile show returned an invalid object", file=sys.stderr)
    raise SystemExit(1)
if not isinstance(plan, dict) or not isinstance(plan.get("steps"), list) or not plan["steps"]:
    print("profile plan returned no steps", file=sys.stderr)
    raise SystemExit(1)
PY
done

printf '%s\n' "$profiles_json" | python3 -c 'import json, sys; json.load(sys.stdin)' >/dev/null
printf 'manifest check: %s schema(s), %s profile(s), %s repository source(s)\n' \
	"$(find schemas -maxdepth 1 -type f -name '*.json' | wc -l)" \
	"$(printf '%s\n' "$profiles_json" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')" \
	"$(printf '%s\n' "$repositories_json" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')"
