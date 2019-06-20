.DEFAULT_GOAL := build
MAKE_PATH=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

build:
	bash $(MAKE_PATH)/build.sh build
	bash $(MAKE_PATH)/../scripts/move-armbian-image.sh

update:
	bash $(MAKE_PATH)/build.sh update
	bash $(MAKE_PATH)/../scripts/move-armbian-image.sh

clean:
	rm -rf $(MAKE_PATH)/armbian/armbian-build
