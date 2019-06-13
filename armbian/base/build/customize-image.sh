#!/bin/bash

set -ex

# This conforms to the expected order of armbian build args
RELEASE=$1
LINUXFAMILY=$2
BOARD=$3
BUILD_DESKTOP=$4

Main() {
        case $RELEASE in
                jessie)
                        exit 1
                        ;;
                xenial)
                        exit 1
                        ;;
                stretch)
                        CustomizeArmbian
                        ;;
                bionic)
                        exit 1
                        ;;
        esac
} # Main

CustomizeArmbian() {
    echo "Running BitBox Base customization script..."

    # copy custom scripts to filesystem
    mkdir -p /opt/shift/scripts
    cp -aR /tmp/overlay/scripts /opt/shift
    chmod -R +x /opt/shift/scripts
    
    # copy configuration items to filesystem
    cp -aR /tmp/overlay/config /opt/shift

    # copy built Go binaries and their associated .service files to filesystem
    cp -aR /tmp/overlay/build/* /opt/shift

    # run our own customization script
    /bin/bash /tmp/overlay/build/customize-armbian-rockpro64.sh
} # CustomizeArmbian

Main "$@"
