#!/bin/bash

# BitBox Base: build Armbian base image
# 
# Script to automate the build process of the customized Armbian base image for the BitBox Base. 
# Additional information: https://digitalbitbox.github.io/bitbox-base
#
set -eu

# Settings
#
# VirtualBox Number of CPU cores
VIRTUALBOX_CPU="4"
# VirtualBox Memory in MB
VIRTUALBOX_MEMORY="8192"

function usage() {
	echo "Build customized Armbian base image for BitBox Base"
	echo "Usage: ${0} [update]"
}

function cleanup() {
	if [[ "${ACTION}" != "clean" ]]; then
		echo "Cleaning up by halting any running vagrant VMs.."
		vagrant halt
	fi
}

ACTION=${1:-"build"}

if ! [[ "${ACTION}" =~ ^(build|update|clean)$ ]]; then
	usage
	exit 1
fi

trap cleanup EXIT

case ${ACTION} in
	build|update)
		if ! which git >/dev/null 2>&1 && ! which vagrant >/dev/null 2>&1; then
			echo
			echo "Build environment not set up, please check documentation at"
			echo "https://digitalbitbox.github.io/bitbox-base"
			echo
			exit 1
		fi

		if [ ! -d "armbian-build" ]; then 
			git clone https://github.com/armbian/build armbian-build
			sed -i "s/#vb.memory = \"8192\"/vb.memory = \"${VIRTUALBOX_MEMORY}\"/g" armbian-build/Vagrantfile
			sed -i "s/#vb.cpus = \"4\"/vb.cpus = \"${VIRTUALBOX_CPU}\"/g" armbian-build/Vagrantfile
			cd armbian-build
		else 
			cd armbian-build
			git pull --no-rebase
		fi

		vagrant up
		mkdir -p output/
		mkdir -p userpatches/overlay
		cp -aR ../base/* userpatches/overlay/					# copy scripts and configuration items to overlay
		cp -aR ../../tools userpatches/overlay/					# copy additional software packages to overlay
		cp -a  ../base/build/customize-image.sh userpatches/	# copy customize script to standard Armbian build hook

		: "${BOARD:=rockpro64}"
		if [ "${ACTION}" == "build" ]; then
			vagrant ssh -c "cd armbian/ && sudo time ./compile.sh BOARD=${BOARD} KERNEL_ONLY=no KERNEL_CONFIGURE=no RELEASE=stretch BRANCH=default BUILD_DESKTOP=no WIREGUARD=no PROGRESS_LOG_TO_FILE=yes LIB_TAG=sunxi-4.20"
		else
			vagrant ssh -c "cd armbian/ && sudo time ./compile.sh BOARD=${BOARD} KERNEL_ONLY=no KERNEL_CONFIGURE=no RELEASE=stretch BRANCH=default BUILD_DESKTOP=no WIREGUARD=no CLEAN_LEVEL="oldcache" PROGRESS_LOG_TO_FILE=yes LIB_TAG=sunxi-4.20"
		fi

		sha256sum output/images/Armbian_*.img
		;;

	clean)
		set +e
		if [ -d "armbian-build" ]; then
			cd armbian-build
			vagrant halt 
			vagrant destroy -f
			cd ..
			rm -rf armbian-build
		fi
		;;

esac
