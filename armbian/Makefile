.DEFAULT_GOAL := build
MAKE_PATH=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

build:
	bash $(MAKE_PATH)/armbian-build.sh build

update:
	bash $(MAKE_PATH)/armbian-build.sh update

mender-artefacts:
	bash $(MAKE_PATH)/armbian-mender-convert.sh

build-ondevice:
	bash $(MAKE_PATH)/armbian-build.sh ondevice

clean:
	rm -rf $(MAKE_PATH)/armbian/armbian-build
