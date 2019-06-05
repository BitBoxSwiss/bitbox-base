.DEFAULT_GOAL=default
HAS_DOCKER := $(shell which docker 2>/dev/null)
REPO_ROOT=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

check-docker:
ifndef HAS_DOCKER
	$(error "This command requires Docker.")
endif

default:
	@mkdir -p $(REPO_ROOT)/build
	@echo "Building tools.."
	$(MAKE) -C tools
	@echo "Building middleware.."
	$(MAKE) -C middleware
	@echo "Building armbian.."
	$(MAKE) -C armbian

dockerinit: check-docker
	docker build --tag digitalbitbox/bitbox-base .

docker-build: dockerinit
	docker run \
	       --rm \
	       --tty \
	       -v $(REPO_ROOT):/opt/go/src/github.com/digitalbitbox/bitbox-base \
	  digitalbitbox/bitbox-base bash -c " \
	      make -C tools && \
	      make -C middleware"

ci: dockerinit
	./scripts/travis-ci.sh
