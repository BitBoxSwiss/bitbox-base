.DEFAULT_GOAL=build-all
HAS_DOCKER := $(shell which docker 2>/dev/null)
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

check-docker:
ifndef HAS_DOCKER
	$(error "This command requires Docker.")
endif

build-go:
	@echo "Building tools.."
	$(MAKE) -C tools
	@echo "Building middleware.."
	$(MAKE) -C middleware

build-all: docker-build-go
	@echo "Building Armbian.."
	$(MAKE) -C armbian

build-update: docker-build-go
	@echo "Updating Armbian build.."
	$(MAKE) update -C armbian

mender-artefacts:
	@echo "Creating Mender update artefacts.."
	$(MAKE) mender-artefacts -C armbian

clean:
	$(MAKE) -C armbian clean
	bash $(REPO_ROOT)/scripts/clean.sh
	# Note that this only delete the final image, not the docker cache
	# You should never need it, but if you want to delete the cache, you can run
	# "docker rmi $(docker images -a --filter=dangling=true -q)"
	docker rmi digitalbitbox/bitbox-base

dockerinit: check-docker
	docker build --tag digitalbitbox/bitbox-base .

docker-build-go: dockerinit
	@echo "Building tools and middleware inside Docker container.."
	docker build --tag digitalbitbox/bitbox-base .
	docker run \
	       --rm \
	       --tty \
	       -v $(REPO_ROOT)/build:/opt/build_host \
	  digitalbitbox/bitbox-base bash -c "cp -f /opt/build/* /opt/build_host"

ci: dockerinit
	./scripts/travis-ci.sh

docker-jekyll: dockerinit
	# TODO(hkjn): Investigate why we need the 'rm -rf', or else Jekyll throws errors like the
	# following, seemingly trying to read a non-existing 'share' file under armbian/armbian-build/packages/bsp/common/usr/
	#   No such file or directory @ realpath_rec - /srv/jekyll/armbian/armbian-build/packages/bsp/common/usr/share
	rm -rf armbian/armbian-build/packages/bsp/common/usr/
	docker run --rm -it -p 4000:4000 -v $(REPO_ROOT):/srv/jekyll jekyll/jekyll:pages jekyll serve --watch --incremental
