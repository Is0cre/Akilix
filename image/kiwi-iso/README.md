# PenSUSE M0 ISO definition

This KIWI definition produces an openSUSE Leap 16 ISO suitable for a first
USB/Ventoy boot test. It intentionally contains platform prerequisites only;
it is not a security-tool distribution.

Build on a privileged Leap 16 build host:

```sh
make kiwi-iso
```

The target builds the current CLI and shell completions into the image overlay
before KIWI runs. The result is written under `build/kiwi-iso/`. Verify the
generated checksum before copying the ISO to Ventoy media. A Ventoy boot test
must verify that the system reaches the expected target, does not start PenSUSE
listeners, and preserves the documented passive-opening behavior.
