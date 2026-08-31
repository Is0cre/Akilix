#!/usr/bin/env python3
import re, sys, xml.etree.ElementTree as ET
from pathlib import Path
from PIL import Image
root=Path(__file__).resolve().parents[1]; fails=[]
expected='source/pensuse-master.png source/pensuse-mark-master.png source/pensuse-wordmark.svg web/pensuse-horizontal.png web/pensuse-horizontal@2x.png web/pensuse-horizontal-dark.png web/pensuse-horizontal-light.png web/pensuse-mark.png web/favicon.svg web/favicon.ico web/favicon-16.png web/favicon-32.png web/favicon-48.png web/apple-touch-icon.png web/android-chrome-192.png web/android-chrome-512.png web/github-social-1280x640.png web/social-card-1200x630.png desktop/pensuse-16.png desktop/pensuse-22.png desktop/pensuse-24.png desktop/pensuse-32.png desktop/pensuse-48.png desktop/pensuse-64.png desktop/pensuse-128.png desktop/pensuse-256.png desktop/pensuse-512.png os/plymouth/logo.png os/plymouth/splash-1920x1080.png os/grub/logo.png os/grub/background-1920x1080.png os/installer/logo.png os/installer/banner.png os/wallpaper/pensuse-1920x1080.png os/wallpaper/pensuse-2560x1440.png os/wallpaper/pensuse-3840x2160.png terminal/pensuse-ascii-full.txt terminal/pensuse-ascii-medium.txt terminal/pensuse-ascii-small.txt terminal/pensuse-ascii-micro.txt terminal/pensuse-unicode.txt terminal/pensuse-ansi.txt terminal/pensuse-motd.txt print/pensuse-sticker.png print/pensuse-sticker.svg print/pensuse-mono-black.svg print/pensuse-mono-white.svg preview.png'.split()
for f in expected:
 p=root/f
 if not p.is_file() or p.stat().st_size==0: fails.append('missing/empty '+f)
for p in root.rglob('*.svg'):
 try: s=p.read_text(); ET.fromstring(s)
 except Exception as e: fails.append(f'SVG {p}: {e}'); continue
 if 'viewBox=' not in s: fails.append('missing viewBox '+str(p))
 if re.search(r'(?:href|xlink:href|src)\s*=\s*["\'](?:https?://|data:)|<script',s,re.I): fails.append('forbidden SVG content '+str(p))
for p in (root/'terminal').glob('pensuse-ascii-*.txt'):
 if any(ord(c)>127 for c in p.read_text()): fails.append('non-ASCII '+p.name)
ansi=(root/'terminal/pensuse-ansi.txt').read_text() if (root/'terminal/pensuse-ansi.txt').exists() else ''
if re.search(r'\x1b\[(?![0-9;]*m)',ansi): fails.append('ANSI contains non-SGR escape')
dims={'web/pensuse-horizontal.png':(1200,420),'web/pensuse-horizontal@2x.png':(2400,840),'web/github-social-1280x640.png':(1280,640),'web/social-card-1200x630.png':(1200,630),'os/plymouth/splash-1920x1080.png':(1920,1080),'os/grub/background-1920x1080.png':(1920,1080),'os/wallpaper/pensuse-1920x1080.png':(1920,1080),'os/wallpaper/pensuse-2560x1440.png':(2560,1440),'os/wallpaper/pensuse-3840x2160.png':(3840,2160)}
for f, expected_size in dims.items():
 try:
  im=Image.open(root/f)
  if im.size!=expected_size: fails.append(f'wrong dimensions {f}: {im.size}')
 except Exception as e: fails.append(f'invalid image {f}: {e}')
try:
 im=Image.open(root/'web/favicon.ico')
 if len(getattr(im,'ico',None).entry if hasattr(im,'ico') else []) < 3: fails.append('favicon.ico missing resolutions')
except Exception as e: fails.append(f'invalid favicon.ico: {e}')
print('PenSUSE branding validation: PASS' if not fails else 'PenSUSE branding validation: FAIL')
for f in fails: print('  FAIL '+f)
sys.exit(1 if fails else 0)
