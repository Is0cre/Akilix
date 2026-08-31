#!/usr/bin/env python3
from pathlib import Path
root = Path(__file__).resolve().parents[1] / 'terminal'; root.mkdir(exist_ok=True)
assets = {
 'akilix-ascii-full.txt': '                         __\n                    _.-\'  `-._\n              _..--\'  _      `-._\n       ______/      _/ \\_         `\\\n   _.-\'     /      /     \\          |\n _/        /      /       \\        /\n/         /      /         \\___.-\'\n\\__      /      /                 \\_\n   `----\'      /____________________/\n        \\_____/       o---o---o\n\nAkilix\nSecurity work with provenance.',
 'akilix-ascii-medium.txt': '             __\n       _..--\'  `--._\n  ____/  _      _   `\\\n /    \\/ \\____/ \\    |\n \\___ /          \\___/\n     `----o-o-o----\n     Akilix',
 'akilix-ascii-small.txt': '      __\n _..-\'  `-._\n/___  ___  _\\\n   `-o-o-o-\'\n Akilix',
 'akilix-ascii-micro.txt': '_/\\_ o-o\nAkilix',
 'akilix-unicode.txt': '    ╭─╮   ●─●─●\n ╭──╯ ╰──╮  Akilix\n ╰───────╯',
 'akilix-ansi-dark.txt': '\033[32m    __/\\__\033[0m  \033[33m●─●─●\033[0m\n\033[1;37mAki\033[32mlix\033[0m\n\033[90mSecurity work with provenance.\033[0m',
 'akilix-ansi-light.txt': '\033[32m    __/\\__\033[0m  \033[33m●─●─●\033[0m\n\033[1;30mAki\033[32mlix\033[0m\n\033[90mSecurity work with provenance.\033[0m',
 'akilix-ansi.txt': '\033[32m       __\033[0m\n\033[32m _..--\'  `-._\033[0m\n\033[1;37mAki\033[32mlix\033[0m  \033[33m●─●─●\033[0m\n\033[90mSecurity work with provenance.\033[0m',
 'akilix-motd.txt': '  __/\\__  ●─●─●\n Akilix\n Security work with provenance.'}
for name, data in assets.items(): (root / name).write_text(data + '\n', encoding='utf-8')
