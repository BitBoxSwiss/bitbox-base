---
layout: default
title: Confgen
nav_order: 125
parent: Custom applications
---
## BitBox Base Confgen

Application to generate text files from a template, replacing placeholders that specify Redis keys.
It's written in Go as part of the [BitBox Base](https://github.com/digitalbitbox/bitbox-base) project by [Shift Cryptosecurity](https://shiftcrypto.ch) and used for automatically generating configuration files.

The program reads a text file specified using the `--template` argument, parses the contents and writes it into a target file that is either specified with the `--output` argument, or directly on the first line of the template.

Helpful in configuration management to generate config files. Multiple fallback options are availabe to ensure resilient creation of valid output.

{% raw %}
```
$ bbbconfgen --help

bbbconfgen version 0.1
generates text files from a template, substituting placeholders with Redis values

Command-line arguments:
  --template      input template text file
  --output        output text file
  --redis-addr    redis connection address (default "localhost:6379")
  --redis-db      redis database number
  --redis-pass    redis password
  --verbose
  --version
  --help

Optionally, the output file can be specified on the first line in the template text file.
This line will be dropped and only used if no --output argument is supplied.
  
  {{ #output: /tmp/output.txt }}
  
Placeholders in the template text file are defined as follows.
Make sure to respect spaces between arguments.

  {{ key }}                     is replaced by Redis 'key', only if key is present
  {{ key #rm }}                 ...deletes the placeholder if key not found
  {{ key #rmLine }}             ...deletes the whole line if key not found
  {{ key #default: some val }}  ...uses default value if key not found
```
{% endraw %}

Check out our own configuration templates to get started: [`/armbian/base/config/templates/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/config/templates)

[See Docs on GitHub](https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbfancontrol){: .btn }
