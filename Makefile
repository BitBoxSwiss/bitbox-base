.DEFAULT_GOAL=build-all
HAS_DOCKER := $(shell which docker 2>/dev/null)
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
PYTHON_CI_IMAGE_VERSION=0.1.0

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
	@echo "Building armbian.."
	$(MAKE) -C armbian

clean:
	$(MAKE) -C armbian clean
	bash $(REPO_ROOT)/scripts/clean.sh
	# Note that this only delete the final image, not the docker cache
	# You should never need it, but if you want to delete the cache, you can run
	# "docker rmi $(docker images -a --filter=dangling=true -q)"
	docker rmi digitalbitbox/bitbox-base

ci: build-docker-ci-image
	./scripts/travis-ci.sh

build-docker-image: check-docker
	docker build --tag digitalbitbox/bitbox-base -f scripts/Dockerfile .

build-docker-ci-image: check-docker
	docker build --tag base-ci:$(PYTHON_CI_IMAGE_VERSION) -f scripts/Dockerfile-python-ci .

docker-build-go: build-docker-image
	@echo "Building tools and middleware inside Docker container.."
	docker run \
	       --rm \
	       --tty \
	       -v $(REPO_ROOT)/build:/opt/build_host \
	  digitalbitbox/bitbox-base bash -c "cp -f /opt/build/* /opt/build_host"

docker-jekyll: build-docker-image
	# TODO(hkjn): Investigate why we need the 'rm -rf', or else Jekyll throws errors like the
	# following, seemingly trying to read a non-existing 'share' file under armbian/armbian-build/packages/bsp/common/usr/
	#   No such file or directory @ realpath_rec - /srv/jekyll/armbian/armbian-build/packages/bsp/common/usr/share
	rm -rf armbian/armbian-build/packages/bsp/common/usr/
	docker run --rm -it -p 4000:4000 -v $(REPO_ROOT):/srv/jekyll jekyll/jekyll:pages jekyll serve --watch --incremental

python-style-check: build-docker-ci-image
	docker run \
	       --rm \
	       --tty \
	       -v $(REPO_ROOT)/:/opt/repo_host \
	       base-ci:$(PYTHON_CI_IMAGE_VERSION)
