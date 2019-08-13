#!/bin/sh

# BitBox Base: mass import key/value pairs into Redis
# 
# Pipe a text file with Redis commands (one per line) to this script to
# convert it to the Redis protocol for mass insertion.
# https://redis.io/topics/mass-insert
#
# Example: cat redis-commands.txt | sh redis-pipe.sh | redis-cli --pipe

while read -r CMD; do
  # each command begins with *{number arguments in command}\r\n
  XS="${CMD}"
  # shellcheck disable=SC2086
  set -- ${XS}
  printf "*%s\r\n" "${#}"
  # for each argument, we append ${length}\r\n{argument}\r\n
  for X in $CMD; do
    printf "\$%s\r\n%s\r\n" "${#X}" "${X}"
  done
done
