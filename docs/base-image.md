---
layout: default
title: Base Image
nav_order: 300
has_children: true
permalink: /base-image
---
## BitBox Base: Creating the Base image

This document provides a high-level overview of the overall build process.
The main output of the build is an Armbian image that can be used to boot the BitBox Base.
See also [Building the Armbian base image](/os/armbian-build.md) and related pages for more details.
The build process contains the following steps, all of which can be initiated by running `make` in the root of the repository.

The steps involved are:

1. build the Go applications
1. build the Armbian image

Unpacking the steps above, what happens when you type `make` is:

1. [`make dockerinit`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile#L28): builds the `digitalbitbox/bitbox-base` image defined by the [`Dockerfile`](https://github.com/digitalbitbox/bitbox-base/blob/master/Dockerfile)
1. [`make docker-build-go`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile#L31): depends on `make dockerinit`, and performs the compilation of the Go applications inside a Docker container from the `digitalbitbox/bitbox-base` image
    1. [`cd tools && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/Makefile#L38): inside the `digitalbitbox/bitbox-base` container, all Go binaries under the `tools/`  subdirectory are built
        1. [`cd bbbfancontrol && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/bbbfancontrol/Makefile): compiles the `build/bbbfancontrol` binary
        1. [`cd bbbsupervisor && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/supervisor/Makefile): compiles the `build/bbbsupervisor` binary
    1. [`cd middleware && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/middleware/Makefile#L39): inside the `digitalbitbox/bitbox-base` container, the `build/base-middleware` binary is built from the `middleware/` subdirectory
1. [`make build-all`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile#L20): the default target if `make` is called, and depends on `make docker-build-go`, and after that performs the main Armbian image build
    1. [`cd armbian && make`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/Makefile): builds the Armbian `.img` file, using the Go binaries in `build/` as inputs
    1. finally, the `.img` file is moved to the `build/` directory

### Overriding user for containerized builds

By default, the `make` commands that use Docker to perform the builds will use a user with the same id as the calling user on the host.
This behavior allows the build to produce outputs to the host `build/` directory without making it owned by the super-user (with id `0`), which could cause permission issues if the non-privileged host user attempts to delete files in the `build/` directory.
Unfortunately, for users who choose to require `sudo` to run `docker` commands, this means that a command like `sudo make` will run as the super-user, so there is no way for the `make` command to find the non-privileged user's id.
As a workaround, such users can either:

1. run `make` and similar commands as usual, and deal with `build/` being owned by the super-user, i.e user with id `0`
1. specify the non-privileged user explicitly when calling `make`:

```bash
sudo make BUILDER_UID=$(id -u)
```

Since the `id -u` is invoked before `sudo` has switched to the super-user, it will return the non-privileged user's id, making it available to the `Makefile`.
