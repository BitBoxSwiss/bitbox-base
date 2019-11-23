---
layout: default
title: Confgen
nav_order: 125
parent: Custom applications
---
## BitBoxBase Confgen

Application to generate text files from a template, replacing placeholders that specify Redis keys.
It's written in Go as part of the [BitBoxBase](https://github.com/digitalbitbox/bitbox-base) project by [Shift Cryptosecurity](https://shiftcrypto.ch) and used for automatically generating configuration files.

The program reads a text file specified using the `--template` argument, parses the contents and writes it into a target file that is either specified with the `--output` argument, or directly on the first line of the template.

Helpful in configuration management to generate config files. Multiple fallback options are availabe to ensure resilient creation of valid output.

```
$ bbbconfgen --help

bbbconfgen version 1.0
generates configuration files from a template, substituting placeholders with Redis values

Command-line arguments:
  --template      input template config file
  --output        output config file
  --redis-addr    redis connection address  (default "localhost:6379")
  --redis-db      redis database number     (default 0)
  --redis-pass    redis password
  --version
  --quiet
  --help

Optionally, the output file can be specified on the first line in the template file.
This line will be dropped and only used if no --output argument is supplied.

  {{ #output: /tmp/output.conf }}

Placeholders in the template file are defined as follows.
Make sure to respect spaces between arguments.

  {{ key }}                     is replaced by Redis 'key', only if key is present
  {{ key #rm }}                 ...deletes the placeholder if key not found
  {{ key #rmLine }}             ...deletes the whole line if key not found
  {{ key #default: some val }}  ...uses default value if key not found

The #rmLineTrue and #rmLineFalse functions allows to drop a line conditionally.

  {{ key #rmLineTrue }}         drop line if a key is set to '1', 'true', 'yes' or 'y'
  {{ key #rmLineFalse }}        drop line if a key is set to '0', 'false', 'no', 'n' or not at all
```

Check out our own configuration templates to get started: [`/armbian/base/config/templates/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/config/templates)

[See Docs on GitHub](https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbfancontrol){: .btn }
