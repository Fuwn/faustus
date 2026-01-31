_faustus() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=($(compgen -W "--help --version -h -v" -- "$cur"))
}

complete -F _faustus faustus
