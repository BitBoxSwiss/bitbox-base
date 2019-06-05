default:
	mkdir -p build
	echo "Building tools.."
	$(MAKE) -C tools
	echo "Building middleware.."
	$(MAKE) -C middleware
	echo "Building armbian.."
	$(MAKE) -C armbian

ci:
	./scripts/travis-ci.sh
