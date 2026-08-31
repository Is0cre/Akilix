# Akilix Shell

The live operator account uses Zsh by default. Akilix builds on openSUSE's
packaged `/etc/zshrc` and adds a small system-wide baseline through the native
`/etc/zsh.zshrc.local` hook. It does not install a plugin framework or perform
network access while starting a shell.

New accounts receive `/etc/skel/.zshrc` with conservative interactive defaults
and a clear place for personal aliases. System behavior is deliberately kept in
`/etc/zsh.zshrc.local`, so completion and history also work when an existing
operator has no personal `.zshrc`.

Interactive history is stored in `~/.zsh_history`, uses Zsh's extended format
with timestamps and durations, and is synchronized between open terminals.
The default retains up to 75,000 entries on disk. Commands beginning with a
space are not recorded, giving operators an explicit way to avoid retaining a
command containing a credential or other sensitive value.

Shell history is operator convenience, not canonical provenance. Commands run
through `akilix run` must continue to produce structured invocation records;
history must never be parsed as a substitute for those records.

Users can override the system baseline in `~/.zshrc`. The image installs the
generated `akilix` completion function through Zsh's standard `fpath` and
initializes Zsh `compinit` from the system baseline. Completion indexes only
locally installed functions and does not contact the network.
