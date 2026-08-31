#!/usr/bin/env python3
"""Offline structural validation for canonical PenSUSE branding inputs."""

import re
import struct
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "branding"
FAILURES = []

REQUIRED = [
    "LICENSE",
    "README.md",
    "source/pensuse-master.png",
    "source/pensuse-mark-master.png",
    "source/pensuse-wordmark.svg",
    "os/grub/background-1920x1080.png",
    "os/grub/logo.png",
    "os/plymouth/logo.png",
    "os/plymouth/splash-1920x1080.png",
    "os/installer/logo.png",
    "os/installer/banner.png",
    "os/wallpaper/pensuse-1920x1080.png",
    "os/wallpaper/pensuse-2560x1440.png",
    "os/wallpaper/pensuse-3840x2160.png",
    "terminal/pensuse-ascii-small.txt",
    "terminal/pensuse-motd.txt",
    "web/favicon.svg",
    "web/favicon.ico",
    "web/pensuse-horizontal.png",
    "web/pensuse-mark.png",
]

PNG_DIMENSIONS = {
    "os/grub/background-1920x1080.png": (1920, 1080),
    "os/grub/logo.png": (420, 420),
    "os/plymouth/logo.png": (512, 512),
    "os/plymouth/splash-1920x1080.png": (1920, 1080),
    "os/installer/logo.png": (800, 280),
    "os/installer/banner.png": (1400, 360),
    "os/wallpaper/pensuse-1920x1080.png": (1920, 1080),
    "os/wallpaper/pensuse-2560x1440.png": (2560, 1440),
    "os/wallpaper/pensuse-3840x2160.png": (3840, 2160),
    "web/pensuse-horizontal.png": (1200, 420),
    "web/pensuse-mark.png": (900, 900),
}


def png_dimensions(path):
    with path.open("rb") as stream:
        header = stream.read(24)
    if len(header) != 24 or header[:8] != b"\x89PNG\r\n\x1a\n" or header[12:16] != b"IHDR":
        raise ValueError("invalid PNG header")
    return struct.unpack(">II", header[16:24])


for relative in REQUIRED:
    path = ROOT / relative
    if not path.is_file() or path.stat().st_size == 0:
        FAILURES.append(f"missing or empty: {relative}")

for relative, expected in PNG_DIMENSIONS.items():
    try:
        actual = png_dimensions(ROOT / relative)
        if actual != expected:
            FAILURES.append(f"wrong dimensions: {relative}: {actual}, want {expected}")
    except (OSError, ValueError) as error:
        FAILURES.append(f"invalid PNG: {relative}: {error}")

for path in ROOT.rglob("*.svg"):
    try:
        text = path.read_text(encoding="utf-8")
        ET.fromstring(text)
    except (OSError, UnicodeError, ET.ParseError) as error:
        FAILURES.append(f"invalid SVG: {path.relative_to(ROOT)}: {error}")
        continue
    if "viewBox=" not in text:
        FAILURES.append(f"SVG lacks viewBox: {path.relative_to(ROOT)}")
    if re.search(r"<script|(?:href|xlink:href|src)\s*=\s*['\"](?:https?://|data:)", text, re.I):
        FAILURES.append(f"SVG contains script or remote/embedded reference: {path.relative_to(ROOT)}")

for path in (ROOT / "terminal").glob("pensuse-ascii-*.txt"):
    try:
        if any(ord(character) > 127 for character in path.read_text(encoding="utf-8")):
            FAILURES.append(f"non-ASCII character in {path.relative_to(ROOT)}")
    except (OSError, UnicodeError) as error:
        FAILURES.append(f"invalid terminal asset: {path.relative_to(ROOT)}: {error}")

if FAILURES:
    print("branding check: FAIL")
    for failure in FAILURES:
        print(f"  {failure}")
    sys.exit(1)

print(f"branding check: {len(REQUIRED)} required asset(s), {len(PNG_DIMENSIONS)} dimension check(s)")
