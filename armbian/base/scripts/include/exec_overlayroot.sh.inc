# function to execute command, either within overlayroot-chroot or directly
#

# check if overlayroot is enabled
OVERLAYROOT_ENABLED=0
if grep -q "tmpfs" /etc/overlayroot.local.conf 2>/dev/null; then
    OVERLAYROOT_ENABLED=1
fi

exec_overlayroot() {
    if [[ "${1}" != "base-only" ]] && [[ "${1}" != "all-layers" ]]; then
        echo "exec_overlayroot(): first argument '${1}', but must be either"
        echo "                    'base-only':  execute base layer (in r/o partition when overlayroot active, or directy when no overlayroot active"
        echo "                    'all-layers': execute both in overlayroot and directly"
        exit 1
    fi

    if [ "${OVERLAYROOT_ENABLED}" -eq 1 ]; then
        echo "executing in overlayroot-chroot: ${2}"
        overlayroot-chroot /bin/bash -c "${2}"
    fi

    if [ "${OVERLAYROOT_ENABLED}" -ne 1 ] || [[ "${1}" == "all-layers" ]]; then
        echo "executing directly: ${2}"
        /bin/bash -c "${2}"
    fi
}
