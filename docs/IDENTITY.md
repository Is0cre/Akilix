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

The Git remote remains at its existing repository location until its owner
renames it. The Go module declares the intended future repository path
`github.com/Is0cre/Akilix`; local builds do not require the remote rename.

## Visual transition

The renamed bitmap files are transitional copies of the earlier artwork and
must be replaced before the next public Akilix ISO. Text, executable names,
desktop integration, boot-theme paths, and image metadata already use Akilix.
Replacement masters must then regenerate and validate every derived bitmap so
no stale wordmark is shipped.
