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
	@echo "Building armbian.."
	$(MAKE) -C armbian

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
