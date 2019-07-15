#!/bin/bash

# BitBox Base: build Armbian base image
#
# Script to automate the build process of the customized Armbian base image for the BitBox Base.
# Additional information: https://digitalbitbox.github.io/bitbox-base
#
set -eu

function usage() {
	echo "Build customized Armbian base image for BitBox Base"
	echo "Usage: ${0} [build|update|ondevice]"
	echo
	echo "running the setup directly ondevice currently support"
	echo "Armbian releases Debian Buster and Ubuntu Bionic"
}

ACTION=${1:-"build"}

if ! [[ "${ACTION}" =~ ^(build|update|ondevice)$ ]]; then
	usage
	exit 1
fi

case ${ACTION} in
	build|update)
		if ! which git >/dev/null 2>&1 || ! which docker >/dev/null 2>&1; then
			echo
			echo "Build environment not set up, please check documentation at"
			echo "https://digitalbitbox.github.io/bitbox-base"
			echo
			exit 1
		fi

		git log --pretty=format:'%h' -n 1 > ./base/config/latest_commit

		if [ ! -d "armbian-build" ]; then
			git clone https://github.com/armbian/build armbian-build
		fi
		cd armbian-build

		mkdir -p output/
		mkdir -p userpatches/overlay/bin
		cp -aR ../base/* userpatches/overlay/					# copy scripts and configuration items to overlay
		cp -aR ../../build/* userpatches/overlay/bin			# copy additional software binaries to overlay
		cp -a  ../base/build/customize-image.sh userpatches/	# copy customize script to standard Armbian build hook

		BOARD=${BOARD:-rockpro64}
		BUILD_ARGS="docker BOARD=${BOARD} KERNEL_ONLY=no KERNEL_CONFIGURE=no RELEASE=bionic BRANCH=default BUILD_DESKTOP=no WIREGUARD=no"
		if ! [ "${ACTION}" == "build" ]; then
			BUILD_ARGS="${BUILD_ARGS} CLEAN_LEVEL=oldcache PROGRESS_LOG_TO_FILE=yes"
		fi
		time ./compile.sh ${BUILD_ARGS}
		;;

	ondevice)
    	# copy custom scripts to filesystem
    	mkdir -p /opt/shift
    	cp -aR base/scripts /opt/shift
    	chmod -R +x /opt/shift/scripts
    
    	# copy configuration items to filesystem
    	cp -aR base/config /opt/shift

    	# copy built Go binaries and their associated .service files to filesystem
    	cp -aR base/build /opt/shift

		base/build/customize-armbian-rockpro64.sh ondevice
esac
