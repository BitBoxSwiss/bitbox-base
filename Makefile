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

dockerinit: check-docker
	docker build --tag digitalbitbox/bitbox-base .

docker-build-go: dockerinit
	@echo "Building tools and middleware inside Docker container.."
	docker run \
	       --rm \
	       --tty \
	       -v $(REPO_ROOT):/opt/go/src/github.com/digitalbitbox/bitbox-base \
	  digitalbitbox/bitbox-base bash -c " \
	      make -C tools && \
	      make -C middleware"

ci: dockerinit
	./scripts/travis-ci.sh
