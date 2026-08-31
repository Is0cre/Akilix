#!/usr/bin/env python3
from pathlib import Path
root = Path(__file__).resolve().parents[1] / 'terminal'; root.mkdir(exist_ok=True)
assets = {
 'pensuse-ascii-full.txt': '                         __\n                    _.-\'  `-._\n              _..--\'  _      `-._\n       ______/      _/ \\_         `\\\n   _.-\'     /      /     \\          |\n _/        /      /       \\        /\n/         /      /         \\___.-\'\n\\__      /      /                 \\_\n   `----\'      /____________________/\n        \\_____/       o---o---o\n\nPenSUSE\nSecurity work with provenance.',
 'pensuse-ascii-medium.txt': '             __\n       _..--\'  `--._\n  ____/  _      _   `\\\n /    \\/ \\____/ \\    |\n \\___ /          \\___/\n     `----o-o-o----\n     PenSUSE',
 'pensuse-ascii-small.txt': '      __\n _..-\'  `-._\n/___  ___  _\\\n   `-o-o-o-\'\n PenSUSE',
 'pensuse-ascii-micro.txt': '_/\\_ o-o\nPenSUSE',
 'pensuse-unicode.txt': '    ╭─╮   ●─●─●\n ╭──╯ ╰──╮  PenSUSE\n ╰───────╯',
 'pensuse-ansi-dark.txt': '\033[32m    __/\\__\033[0m  \033[33m●─●─●\033[0m\n\033[1;37mPen\033[32mSUSE\033[0m\n\033[90mSecurity work with provenance.\033[0m',
 'pensuse-ansi-light.txt': '\033[32m    __/\\__\033[0m  \033[33m●─●─●\033[0m\n\033[1;30mPen\033[32mSUSE\033[0m\n\033[90mSecurity work with provenance.\033[0m',
 'pensuse-ansi.txt': '\033[32m       __\033[0m\n\033[32m _..--\'  `-._\033[0m\n\033[1;37mPen\033[32mSUSE\033[0m  \033[33m●─●─●\033[0m\n\033[90mSecurity work with provenance.\033[0m',
 'pensuse-motd.txt': '  __/\\__  ●─●─●\n PenSUSE\n Security work with provenance.'}
for name, data in assets.items(): (root / name).write_text(data + '\n', encoding='utf-8')
