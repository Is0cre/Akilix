package completion

const Zsh = `#compdef pensuse

_pensuse_workbooks() {
  local -a workbooks
  workbooks=("${(@f)$(pensuse workbook list 2>/dev/null | awk '{print $1}')}")
  if (( ${#workbooks} )); then
    _describe 'workbooks' workbooks
  fi
}

_pensuse() {
  if (( CURRENT >= 4 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == "workbook" ]]; then
    case "$words[3]" in
      open|status|close|reopen|rename) _pensuse_workbooks; return ;;
    esac
  fi
  _arguments -C \
    '1:command:->command' \
    '*::argument:->argument'
  case $state in
    command) _values 'command' \
      'version[Show PenSUSE version]' \
      'workbook[Manage workbooks]' \
      'scope[Manage scope]' \
      'evidence[Manage evidence]' \
      'run[Execute and inspect invocations]' \
      'container[Inspect container images]' \
      'completion[Generate shell completion]' ;;
    argument)
      case $words[2] in
        workbook) _values 'workbook command' create list open status close reopen rename; [[ $CURRENT -ge 4 ]] && _pensuse_workbooks ;;
        scope) _values 'scope command' add remove exclude list check ;;
        evidence) _values 'evidence command' import list verify ;;
        run) _message 'use: pensuse run WORKBOOK -- COMMAND [ARGS...]' ;;
        container) _values 'container command' inspect ;;
        completion) _values 'shell' zsh bash ;;
      esac ;;
  esac
}
_pensuse "$@"
`

const Bash = `_pensuse_complete() {
  local cur prev
  cur="${COMP_WORDS[COMP_CWORD]}"; prev="${COMP_WORDS[COMP_CWORD-1]}"
  if [[ ${COMP_CWORD} -eq 1 ]]; then COMPREPLY=($(compgen -W 'version workbook scope evidence run container completion' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == workbook && ${COMP_CWORD} -eq 2 ]]; then COMPREPLY=($(compgen -W 'create list open status close reopen rename' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == workbook && ${COMP_CWORD} -ge 3 ]]; then local workbooks; workbooks="$(pensuse workbook list 2>/dev/null | awk '{print $1}')"; COMPREPLY=($(compgen -W "$workbooks" -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == scope ]]; then COMPREPLY=($(compgen -W 'add remove exclude list check' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == evidence ]]; then COMPREPLY=($(compgen -W 'import list verify' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == container ]]; then COMPREPLY=($(compgen -W 'inspect' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == completion ]]; then COMPREPLY=($(compgen -W 'zsh bash' -- "$cur")); return; fi
}
complete -F _pensuse_complete pensuse
`
