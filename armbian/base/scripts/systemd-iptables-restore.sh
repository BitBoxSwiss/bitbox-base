#!/bin/bash
# shellcheck disable=SC1091
#
# This script restores the iptables firewall rules
#
set -eu

# flush existing iptables rules
iptables -P INPUT ACCEPT
iptables -P FORWARD ACCEPT
iptables -P OUTPUT ACCEPT
iptables -t nat -F
iptables -t mangle -F
iptables -F
iptables -X

# restore iptables rules
/sbin/iptables-restore < /etc/iptables/iptables.rules

# list active iptables rules
iptables -L -n
