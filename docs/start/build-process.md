---
layout: default
title: Build process
parent: Getting started
nav_order: 100
---
## Build process

*(TODO)Stadicus: extend and include Mender.io*

The build process contains the following steps, all of which can be initiated by running `make` in the root of the repository.

The steps involved are:

1. build the Go applications
2. build the Armbian image

Unpacking the steps above, what happens when you type `make` is:

1. [`make dockerinit`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile#L28): builds the `digitalbitbox/bitbox-base` image defined by the [`Dockerfile`](https://github.com/digitalbitbox/bitbox-base/blob/master/Dockerfile)
1. [`make docker-build-go`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile#L31): depends on `make dockerinit`, and performs the compilation of the Go applications inside a Docker container from the `digitalbitbox/bitbox-base` image
    1. [`cd tools && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/Makefile#L38): inside the `digitalbitbox/bitbox-base` container, all Go binaries under the `tools/`  subdirectory are built
        1. [`cd bbbfancontrol && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/bbbfancontrol/Makefile): compiles the `build/bbbfancontrol` binary
        1. [`cd bbbsupervisor && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/supervisor/Makefile): compiles the `build/bbbsupervisor` binary
    1. [`cd middleware && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/middleware/Makefile#L39): inside the `digitalbitbox/bitbox-base` container, the `build/bbbmiddleware` binary is built from the `middleware/` subdirectory
1. [`make build-all`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile#L20): the default target if `make` is called, and depends on `make docker-build-go`, and after that performs the main Armbian image build
    1. [`cd armbian && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/Makefile): builds the Armbian `.img` file, using the Go binaries in `build/` as inputs
    1. finally, the `.img` file is moved to the `build/` directory
