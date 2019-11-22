.DEFAULT_GOAL=build-all
HAS_DOCKER := $(shell which docker 2>/dev/null)
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

check-docker:
ifndef HAS_DOCKER
	$(error "This command requires Docker.")
endif

# build Go tools locally: bbbmiddleware, bbbsupervisor, bbbfancontrol
# depends on correct golang setup
build-go:
	@echo "Building tools.."
	$(MAKE) -C tools
	@echo "Building middleware.."
	$(MAKE) -C middleware

# initialize docker environment to build Go tools
dockerinit: check-docker
	docker build --tag digitalbitbox/bitbox-base .

# build Go tools in Docker: bbbmiddleware, bbbsupervisor, bbbfancontrol
# depends on docker setup, no local golang necessary
docker-build-go: dockerinit
	@echo "Building tools and middleware inside Docker container.."
	docker build --tag digitalbitbox/bitbox-base .
	docker run \
	       --rm \
	       --tty \
	       -v $(REPO_ROOT)/bin/go:/opt/build_host \
	  digitalbitbox/bitbox-base bash -c "cp -f /opt/build/* /opt/build_host"

# build Armbian disk image
# see configuration: armbian/base/build/build.conf
build-all: docker-build-go
	@echo "Building Armbian.."
	$(MAKE) -C armbian
	$(MAKE) mender-artefacts -C armbian

# build Armbian disk image, use cached binaries and update customization only
# see configuration: armbian/base/build/build.conf
update-all: docker-build-go
	@echo "Updating Armbian build.."
	$(MAKE) update -C armbian
	$(MAKE) mender-artefacts -C armbian

# create a Mender-enabled disk image out of an Armbian image
mender-artefacts:
	@echo "Creating Mender update artefacts.."
	$(MAKE) mender-artefacts -C armbian

# cleanup build environment
clean:
	$(MAKE) -C armbian clean
	bash $(REPO_ROOT)/contrib/clean.sh
	# Note that this only delete the final image, not the docker cache
	# You should never need it, but if you want to delete the cache, you can run
	# "docker rmi $(docker images -a --filter=dangling=true -q)"
	docker rmi digitalbitbox/bitbox-base

# run CI tests
ci: dockerinit
	./contrib/travis-ci.sh

# WIP: build Armbian image using Jekyll
docker-jekyll: dockerinit
	# TODO(hkjn): Investigate why we need the 'rm -rf', or else Jekyll throws errors like the
	# following, seemingly trying to read a non-existing 'share' file under armbian/armbian-build/packages/bsp/common/usr/
	#   No such file or directory @ realpath_rec - /srv/jekyll/armbian/armbian-build/packages/bsp/common/usr/share
	rm -rf armbian/armbian-build/packages/bsp/common/usr/
	docker run --rm -it -p 4000:4000 -v $(REPO_ROOT):/srv/jekyll jekyll/jekyll:pages jekyll serve --watch --incremental
