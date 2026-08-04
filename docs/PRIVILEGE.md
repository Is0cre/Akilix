# PenSUSE Privilege Model

## Principle

Ordinary PenSUSE operation is unprivileged.

Root access is exceptional, explicit, and narrowly scoped.

---

# Unprivileged operations

Examples should include:

- workbook creation/opening
- scope editing
- evidence metadata inspection
- hashing ordinary readable files
- rootless Podman execution
- report drafting
- local indexing
- local LLM analysis where device permissions allow it

---

# Privileged operations

Potential examples:

- raw block-device acquisition
- changing block-device read-only state
- selected firmware inspection
- memory acquisition
- host profile installation
- Btrfs/Snapper system changes
- hardware access requiring elevated permissions

Privilege must not be granted simply because a security tool expects it.

---

# Early milestones

Do not introduce a permanently running privileged PenSUSE daemon in M0 or M1.

Use explicit elevation where necessary.

---

# Future helper

If a privileged helper becomes necessary, it should:

- expose narrow typed operations
- reject arbitrary command execution
- validate all device paths and arguments
- authenticate callers
- log the requested action
- minimize dependencies
- minimize parsing
- fail closed
- have dedicated tests
- be constrained by SELinux where possible

---

# Separation

The main CLI should not become root merely because one sub-operation requires privilege.

Conceptually:

    user pensuse process
          |
          | explicit authorized request
          v
    narrow privileged helper
          |
          v
    specific kernel/device operation

---

# Safety guards vs boundaries

Scope checks and confirmation prompts are safety guards.

Unix permissions, SELinux, namespaces, capabilities, and mount flags may act as security boundaries.

PenSUSE documentation must distinguish these categories.
