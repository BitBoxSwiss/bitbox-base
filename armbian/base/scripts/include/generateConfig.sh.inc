# generates a config file using custom bbbconfig
# argument is template filename, without path
#

# check if overlayroot is enabled
OVERLAYROOT_ENABLED=0
if grep -q "tmpfs" /etc/overlayroot.local.conf; then
    OVERLAYROOT_ENABLED=1
fi

generateConfig() {
    local TEMPLATES_DIR="/opt/shift/config/templates"

    if [ ${#} -eq 0 ] || [ ${#} -gt 1 ]; then
        echo "ERR: generateConfig() expects exactly one argument"
        exit 1
    fi

    # check template
    local FILE="${TEMPLATES_DIR}/${1}"
    if [ -f "${FILE}" ]; then

        # check if overlayroot enabled
        if [ "${OVERLAYROOT_ENABLED}" -eq 1 ]; then
            echo "generateConfig() from ${FILE}, read-only partition of overlayroot"
            /usr/local/sbin/bbbconfgen --template "${FILE}"
            overlayroot-chroot /bin/bash -c "/usr/local/sbin/bbbconfgen --template '${FILE}'"
        else
            echo "generateConfig() from ${FILE}"
            /usr/local/sbin/bbbconfgen --template "${FILE}"
        fi

    else
        echo "ERR: generateConfig() template file ${FILE} not found"
        exit 1
    fi
}
