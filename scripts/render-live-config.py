#!/usr/bin/env python3
"""Render the local-only KIWI live-image password configuration."""

import re
import sys
from pathlib import Path


PLACEHOLDER = "__AKILIX_LIVE_PASSWORD_HASH__"
HASH_PATTERN = re.compile(r"^\$6\$[A-Za-z0-9./$]+$")


def main() -> int:
    if len(sys.argv) != 4:
        print(f"usage: {sys.argv[0]} TEMPLATE OUTPUT SHA512_CRYPT_HASH", file=sys.stderr)
        return 2
    template, output, password_hash = map(Path, sys.argv[1:])
    value = str(password_hash)
    if not HASH_PATTERN.fullmatch(value):
        print("password must be a SHA-512 crypt hash", file=sys.stderr)
        return 2
    source = template.read_text(encoding="utf-8")
    if source.count(PLACEHOLDER) != 1:
        print("live config must contain exactly one password placeholder", file=sys.stderr)
        return 2
    output.write_text(source.replace(PLACEHOLDER, value), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
