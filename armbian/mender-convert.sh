#!/bin/bash

# BitBoxBase: Create Mender-enabled images
#
# Script to automate the conversion process of an Armbian image into a
# Mender provisioning image and update artefacts for the BitBoxBase.
#
set -eu

function usage() {
	echo "Convert Armbian image into Mender image & update artefact"
	echo "Usage: ${0} [envinit|build]"
}

if [ -f .cleanup-loop-devices ]; then
	../contrib/cleanup-loop-devices.sh
fi

ACTION=${1:-"build"}
SOURCE_NAME="BitBoxBase_Armbian_RockPro64"
VERSION="$(head -n1 base/config/version)"
TEMP_NAME="BitBoxBase"
TARGET_NAME="BitBoxBase-v${VERSION}-RockPro64"

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

		# cleanup loop devices if trigger file present
		if [ -f .cleanup-loop-devices ]; then
			../contrib/cleanup-loop-devices.sh
		fi

		# conversion settings
		DEVICE_TYPE="rockpro64"
		RAW_DISK_IMAGE="input/${SOURCE_NAME}.img"
		ARTIFACT_NAME="${TEMP_NAME}"
		MENDER_DISK_IMAGE="${TEMP_NAME}"

		# retrieve latest Armbian image
		if [ ! -f "../../bin/img-armbian/${SOURCE_NAME}.img" ]; then
			echo "Error: Armbian source file 'bin/img-armbian/${SOURCE_NAME}.img' missing."
			exit 1
		fi
		echo "Copying ./bin/img-armbian/${SOURCE_NAME}.img..."
		cp -f "../../bin/img-armbian/${SOURCE_NAME}.img" "input/"

		./docker-mender-convert from-raw-disk-image \
			--raw-disk-image $RAW_DISK_IMAGE \
			--mender-disk-image $MENDER_DISK_IMAGE \
			--device-type $DEVICE_TYPE \
			--artifact-name $ARTIFACT_NAME \
			--bootloader-toolchain aarch64-linux-gnu

		# move converted images and update artefacts to /provisioning
		echo "Cleaning up..."
		rm "input/${SOURCE_NAME}.img"
		rm "output/${TEMP_NAME}.ext4"

		mkdir -p "../../bin/img-mender/${VERSION}"

		cd output
		mv "${TEMP_NAME}.sdimg" "${TARGET_NAME}.img"
		tar -czf "${TARGET_NAME}.tar.gz" "${TARGET_NAME}.img"
		mv "${TEMP_NAME}.mender" "${TARGET_NAME}.base-unsigned"
		mv "${TARGET_NAME}"* "../../../bin/img-mender/${VERSION}/"
		cd ..

		echo "Mender files ready for provisioning:"
		ls -lh "../../bin/img-mender/${VERSION}/${TARGET_NAME}"*
		echo
		echo "Write to eMMC with the following command (check target device /dev/sdX first!):"
		echo "dd if=./bin/img-mender/${VERSION}/${TARGET_NAME}.img of=/dev/sdX bs=4M conv=sync status=progress && sync"
		echo
        ;;
esac

if [ -f .cleanup-loop-devices ]; then
	../contrib/cleanup-loop-devices.sh
fi
