# Akilix interactive-login status. No network access and no hidden monitoring.
case $- in
    *i*) ;;
    *) return 0 ;;
esac
[ -t 1 ] || return 0
[ "${AKILIX_MOTD_SHOWN:-0}" = 0 ] || return 0
export AKILIX_MOTD_SHOWN=1

if [ "${TERM:-dumb}" != dumb ]; then
    accent='\033[38;5;118m'
    amber='\033[38;5;214m'
    reset='\033[0m'
else
    accent=''
    amber=''
    reset=''
fi

# Blink is deliberately opt-in because it is distracting and some terminals
# implement it as bright background. Set AKILIX_MOTD_BLINK=1 to enable it.
if [ "${AKILIX_MOTD_BLINK:-0}" = 1 ] && [ "${TERM:-dumb}" != dumb ]; then
    amber='\033[5;38;5;214m'
fi

printf '%bAkilix%b  %s  %s\n' "$accent" "$reset" "$(uname -r)" "$(hostname)"
printf '%bACQUISITION%b  unknown storage is untrusted; writes require explicit action\n' "$amber" "$reset"
printf 'Start workspace: akilix    Hardware: akilix acquire inspect\n\n'

unset accent amber reset
