DOCKER := $(shell which docker 2> /dev/null)

default:
	mkdir -p build
	echo "Building tools.."
	$(MAKE) -C tools
	echo "Building middleware.."
	$(MAKE) -C middleware
	echo "Building armbian.."
	$(MAKE) -C armbian

docker-jekyll:
ifndef DOCKER
	$(error "This rule requires Docker to run jekyll.")
endif
	@echo "Starting docker-jekyll server at localhost:4000.."
	docker run --rm -it -p 4000:4000 -v $(shell pwd):/srv/jekyll \
	       jekyll/jekyll:pages jekyll serve --watch --incremental

ci:
	./scripts/travis-ci.sh
