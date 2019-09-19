# error handling function

errorExit() {
    echo "$@" 1>&2
    exit 1
}