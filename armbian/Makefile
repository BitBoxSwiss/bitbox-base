.DEFAULT_GOAL := build
MAKE_PATH=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# build Armbian image from scratch, not using cached binaries
build:
	bash $(MAKE_PATH)/armbian-build.sh build

# update Armbian image with cached binaries
update:
	bash $(MAKE_PATH)/armbian-build.sh update

# create Mender.io artefacts from the latest Armbian image
mender-artefacts:
	bash $(MAKE_PATH)/mender-convert.sh

# update Armbian image and create new Mender.io artefacts in one go
update-mender:
	bash $(MAKE_PATH)/armbian-build.sh update
	bash $(MAKE_PATH)/mender-convert.sh

# execute configuration script directly on the device, starting 
# with a stock Armbian image
build-ondevice:
	bash $(MAKE_PATH)/armbian-build.sh ondevice

# clean up build repositories and cached data
clean:
	rm -rf $(MAKE_PATH)/armbian/armbian-build
