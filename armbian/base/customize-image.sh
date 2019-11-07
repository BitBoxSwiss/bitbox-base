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
            CustomizeArmbian
            ;;
    esac
} # Main

CustomizeArmbian() {
    echo "Running BitBoxBase customization script..."

    # copy custom files to Armbian build overlay
    mkdir -p /opt/shift
    cp -aR /tmp/overlay/* /opt/shift
    chmod -R +x /opt/shift/scripts

    # run custom customization script
    /bin/bash /tmp/overlay/customize-armbian-rockpro64.sh
} # CustomizeArmbian

Main "$@"
