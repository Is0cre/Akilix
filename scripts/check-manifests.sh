#!/bin/sh
# Validate repository-owned machine-readable manifests without changing host state.
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

for schema in schemas/*.json; do
	python3 -m json.tool "$schema" >/dev/null
done

# Use the built CLI when available so this check exercises the shipped binary.
# Fall back to go run for direct developer use before `make build`.
cli=./pensuse
if [ ! -x "$cli" ]; then
	cli="go run ./cmd/pensuse"
fi

profile_dir=${PENSUSE_PROFILE_DIR:-profiles}
profiles_json=$(PENSUSE_PROFILE_DIR="$profile_dir" $cli profile list --json)
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
	show_json=$(PENSUSE_PROFILE_DIR="$profile_dir" $cli profile show "$profile_id" --json)
	plan_json=$(PENSUSE_PROFILE_DIR="$profile_dir" $cli profile plan "$profile_id" --json)
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
printf 'manifest check: %s schema(s), %s profile(s)\n' \
	"$(find schemas -maxdepth 1 -type f -name '*.json' | wc -l)" \
	"$(printf '%s\n' "$profiles_json" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')"
