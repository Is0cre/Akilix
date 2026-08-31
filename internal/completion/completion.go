package completion

const Zsh = `#compdef pensuse

_pensuse_workbooks() {
  local -a workbooks
  workbooks=("${(@f)$(pensuse workbook list 2>/dev/null | awk '{print $1}')}")
  if (( ${#workbooks} )); then
    _describe 'workbooks' workbooks
  fi
}

_pensuse_evidence_ids() {
  local -a evidence
  evidence=("${(@f)$(pensuse evidence list "$words[4]" 2>/dev/null | awk '{print $1}')}")
  if (( ${#evidence} )); then _describe 'evidence' evidence; fi
}

_pensuse_images() {
  local -a images
  images=("${(@f)$(podman images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null)}")
  if (( ${#images} )); then _describe 'local images' images; fi
}

_pensuse() {
  if (( CURRENT >= 4 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == "workbook" ]]; then
    case "$words[3]" in
      open|overview|path|status|close|reopen|rename|validate) _pensuse_workbooks; return ;;
    esac
  fi
  if (( CURRENT >= 5 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == "evidence" && "$words[3]" == "verify" ]]; then _pensuse_evidence_ids; return; fi
  if (( CURRENT >= 6 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == "container" && "$words[3]" == "run" ]]; then _values 'options' --target --override --json --workdir --env --; return; fi
  if (( CURRENT >= 5 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == "container" && "$words[3]" == "run" ]]; then _pensuse_images; return; fi
  if (( CURRENT >= 4 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == "container" && "$words[3]" == "run" ]]; then _pensuse_workbooks; return; fi
  if (( CURRENT >= 4 )) && [[ "$words[1]" == "pensuse" && "$words[2]" == (evidence|scope|run) ]]; then
    case "$words[3]" in
      import|list|verify|add|remove|exclude|check) _pensuse_workbooks; return ;;
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
      'logging[Inspect workbook logging policy]' \
      'evidence[Manage evidence]' \
      'run[Execute and inspect invocations]' \
      'container[Inspect container images]' \
      'profile[Inspect capability profiles]' \
      'repository[Inspect repository trust metadata]' \
      'config[Inspect effective configuration]' \
      'completion[Generate shell completion]' ;;
    argument)
      case $words[2] in
        workbook) _values 'workbook command' create list open overview path status close reopen rename validate; [[ $CURRENT -ge 4 ]] && _pensuse_workbooks ;;
        scope) _values 'scope command' add remove exclude list check ;;
        logging) _values 'logging command' status ;;
        evidence) _values 'evidence command' import list verify ;;
        run) _message 'use: pensuse run WORKBOOK -- COMMAND [ARGS...]' ;;
        container) _values 'container command' inspect run ;;
        profile) _values 'profile command' list show plan ;;
        repository) _values 'repository command' list show ;;
        config) _values 'config command' show path ;;
        completion) _values 'shell' zsh bash ;;
      esac ;;
  esac
}
_pensuse "$@"
`

const Bash = `_pensuse_complete() {
  local cur prev
  cur="${COMP_WORDS[COMP_CWORD]}"; prev="${COMP_WORDS[COMP_CWORD-1]}"
  if [[ ${COMP_CWORD} -eq 1 ]]; then COMPREPLY=($(compgen -W 'version workbook scope logging evidence run container profile repository config completion' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == workbook && ${COMP_CWORD} -eq 2 ]]; then COMPREPLY=($(compgen -W 'create list open overview path status close reopen rename validate' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == workbook && ${COMP_CWORD} -ge 3 ]]; then local workbooks; workbooks="$(pensuse workbook list 2>/dev/null | awk '{print $1}')"; COMPREPLY=($(compgen -W "$workbooks" -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == evidence && ${COMP_WORDS[2]} == verify && ${COMP_CWORD} -ge 4 ]]; then local ids; ids="$(pensuse evidence list "${COMP_WORDS[3]}" 2>/dev/null | awk '{print $1}')"; COMPREPLY=($(compgen -W "$ids" -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == container && ${COMP_WORDS[2]} == run && ${COMP_CWORD} -ge 5 ]]; then COMPREPLY=($(compgen -W '--target --override --json --workdir --env --' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == container && ${COMP_WORDS[2]} == run && ${COMP_CWORD} -ge 4 ]]; then local images; images="$(podman images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null)"; COMPREPLY=($(compgen -W "$images" -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == container && ${COMP_WORDS[2]} == run && ${COMP_CWORD} -ge 3 ]]; then local workbooks; workbooks="$(pensuse workbook list 2>/dev/null | awk '{print $1}')"; COMPREPLY=($(compgen -W "$workbooks" -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == evidence || ${COMP_WORDS[1]} == scope || ${COMP_WORDS[1]} == logging || ${COMP_WORDS[1]} == run ]] && [[ ${COMP_CWORD} -ge 3 ]]; then local workbooks; workbooks="$(pensuse workbook list 2>/dev/null | awk '{print $1}')"; COMPREPLY=($(compgen -W "$workbooks" -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == scope ]]; then COMPREPLY=($(compgen -W 'add remove exclude list check' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == logging ]]; then COMPREPLY=($(compgen -W 'status' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == evidence ]]; then COMPREPLY=($(compgen -W 'import list verify' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == container ]]; then COMPREPLY=($(compgen -W 'inspect' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == profile ]]; then COMPREPLY=($(compgen -W 'list show plan' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == repository ]]; then COMPREPLY=($(compgen -W 'list show' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == config ]]; then COMPREPLY=($(compgen -W 'show path' -- "$cur")); return; fi
  if [[ ${COMP_WORDS[1]} == completion ]]; then COMPREPLY=($(compgen -W 'zsh bash' -- "$cur")); return; fi
}
complete -F _pensuse_complete pensuse
`
