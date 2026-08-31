# Start the PenSUSE graphical session only for the live operator on tty1.
# Exiting or failing Sway returns to the console login shell.
if [ "${USER:-}" = "pensuse" ] && [ "${XDG_VTNR:-}" = "1" ] && \
   [ -z "${WAYLAND_DISPLAY:-}" ] && [ -z "${DISPLAY:-}" ] && \
   command -v sway >/dev/null 2>&1; then
	exec sway
fi
