#compdef faustus

_faustus() {
    _arguments \
        '(-h --help)'{-h,--help}'[Show help]' \
        '(-v --version)'{-v,--version}'[Show version]'
}

_faustus "$@"
