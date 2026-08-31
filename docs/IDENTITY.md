# Akilix Identity Migration

Akilix is the canonical project and distribution name. It derives from Arkin,
through Aki Linux, with an intentional echo of Achilles. The canonical command,
RPM, image, runtime directory, schema prefix, and display name are `akilix` or
`Akilix` as appropriate.

This development-stage rename is a clean identity boundary. Pre-release data
written beneath legacy runtime paths or using legacy schema identifiers is not
silently rewritten: forensic and provenance records must never be mutated by a
branding migration. Any future importer must preserve the original bytes and
record the migration explicitly.

The canonical Git remote and Go module path are `github.com/Is0cre/Akilix`.

## Visual transition

All canonical and generated artwork uses the Akilix name. The monitor-lizard
masters are name-neutral source art; reproducible derivatives add the Akilix
wordmark for documentation, web, installer, boot, and desktop surfaces.
