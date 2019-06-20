---
layout: default
title: Building Go binaries
nav_order: 100
parent: Go applications
---
## BitBox Base: Building Go binaries

The BitBox Base runs custom software written in Go that has to be compiled and become part of [the Armbian image](/os/armbian-build.md).
This page describes the process used to build those images.

The top-level [`Makefile`](https://github.com/digitalbitbox/bitbox-base/blob/master/Makefile) for the repository has two targets:

- `make docker-build-go`: Build the Go applications inside a Docker container
- `make build-go`: Build the Go applications on the host

The default `make` target invokes the `make docker-build-go` target to produce Go binaries compiled for the target CPU architecture in the `build/` directory, and then builds [the Armbian image](/os/armbian-build.md), using the `build/` contents as inputs.
For users that have the Go toolchain installed on the host system, `make build-go` should also work fine, and if the Go environment is not configured correctly, the command should produce some useful information to debug the issue.
