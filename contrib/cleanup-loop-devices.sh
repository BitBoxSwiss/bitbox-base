#!/bin/bash
#
# Remove unused loop devices
# 

if [[ ${UID} -ne 0 ]]; then
    echo "${0}: needs to be run as superuser." >&2
    exit 1
fi

echo "Removing unused loop devices:"
for LOOPS in $(losetup -a | grep "(/)\|BitBoxBase" | cut -f 1 -d ":"); do
    echo "- ${LOOPS}"
    losetup -d "${LOOPS}";
done

dmsetup remove_all
