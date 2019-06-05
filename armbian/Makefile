MAKE_PATH=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

default:
	bash $(MAKE_PATH)/build.sh build

update:
	bash $(MAKE_PATH)/build.sh update

clean:
	bash $(MAKE_PATH)/build.sh clean
