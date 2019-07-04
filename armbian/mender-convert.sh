#!/bin/bash

# BitBox Base: Create Mender-enabled images
# 
# Script to automate the conversion process of an Armbian image into a
# Mender provisioning image and update artefacts for the BitBox Base. 
#
set -eu

function usage() {
	echo "Convert Armbian image into Mender image & update artefact"
	echo "Usage: ${0} [envinit|build]"
}

ACTION=${1:-"build"}
SOURCE_NAME="BitBoxBase_Armbian_RockPro64"
TARGET_NAME="${SOURCE_NAME}-`date +%Y%m%d`"

if ! [[ "${ACTION}" =~ ^(envinit|build)$ ]]; then
	usage
	exit 1
fi

case ${ACTION} in
	build)
		# initialize conversion environment
		if [ ! -d "mender-convert" ]; then
			git clone https://github.com/mendersoftware/mender-convert
			cd mender-convert
			./docker-build arm64
			mkdir -p input
		else
			cd mender-convert
		fi

		# conversion settings
		DEVICE_TYPE="rockpro64"
		RAW_DISK_IMAGE="input/${SOURCE_NAME}.img"
		ARTIFACT_NAME="${TARGET_NAME}"
		MENDER_DISK_IMAGE="${TARGET_NAME}.img"

		# retrieve latest Armbian image
		if [ ! -f "../../provisioning/${SOURCE_NAME}.img" ]; then
			echo "Error: Armbian source file 'provisioning/${SOURCE_NAME}.img' missing."
			exit 1
		fi
		cp -f "../../provisioning/${SOURCE_NAME}.img" "input/"

		./docker-mender-convert from-raw-disk-image \
			--raw-disk-image $RAW_DISK_IMAGE \
			--mender-disk-image $MENDER_DISK_IMAGE \
			--device-type $DEVICE_TYPE \
			--artifact-name $ARTIFACT_NAME \
			--bootloader-toolchain aarch64-linux-gnu

		# move converted images and update artefacts to /provisioning
		rm "input/${SOURCE_NAME}.img"
		mv output/${SOURCE_NAME}* ../../provisioning/
        ;;
esac
