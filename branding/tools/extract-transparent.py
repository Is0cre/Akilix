#!/usr/bin/env python3
"""Remove a baked neutral checkerboard from generated reference artwork.

Only edge-connected near-neutral light pixels are removed; enclosed bone
highlights remain part of the illustration.
"""
from collections import deque
from pathlib import Path
from PIL import Image
import sys

src, dst = map(Path, sys.argv[1:])
im = Image.open(src).convert('RGBA')
px = im.load(); w, h = im.size
def bg(x, y):
    r, g, b, a = px[x, y]
    return a > 0 and max(r, g, b) - min(r, g, b) <= 12 and min(r, g, b) >= 218
seen = bytearray(w*h); q = deque()
for x in range(w): q.extend(((x, 0), (x, h-1)))
for y in range(h): q.extend(((0, y), (w-1, y)))
while q:
    x, y = q.popleft(); i = y*w+x
    if seen[i] or not bg(x, y): continue
    seen[i] = 1
    for nx, ny in ((x-1,y),(x+1,y),(x,y-1),(x,y+1)):
        if 0 <= nx < w and 0 <= ny < h and not seen[ny*w+nx]: q.append((nx, ny))
for i, marked in enumerate(seen):
    if marked: px[i % w, i // w] = (0, 0, 0, 0)
im.save(dst, 'PNG', optimize=False)
