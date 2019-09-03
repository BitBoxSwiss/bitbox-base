#!/bin/bash
#
# This script is run by systemd using the ExecStartPost option 
# after starting bitcoind.service (Bitcoin Core).
set -eu

# We set rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie, but there seems to be
# no way to specify where to expect the bitcoin cookie for c-lightning, so let's
# create a symlink at the expected testnet location.
mkdir -p /mnt/ssd/bitcoin/.bitcoin/testnet3/
ln -fs /mnt/ssd/bitcoin/.bitcoin/.cookie /mnt/ssd/bitcoin/.bitcoin/testnet3/.cookie
echo "${0}: symlink from file .bitcoin/.cookie -> .bitcoin/testnet3/.cookie created."

# wait a few seconds before providing cookie authentication 
# as .env file for electrs and bbbmiddleware 
sleep 10
echo -n 'RPCPASSWORD=' > /mnt/ssd/bitcoin/.bitcoin/.cookie.env
tail -c +12 /mnt/ssd/bitcoin/.bitcoin/.cookie >> /mnt/ssd/bitcoin/.bitcoin/.cookie.env
echo "${0}: file /mnt/ssd/bitcoin/.bitcoin/.cookie.env updated."

# log bitcoind restarts including auth information
echo "`date +%Y-%m-%d-%H:%M` systemd-bitcoind-post.sh `cat /mnt/ssd/bitcoin/.bitcoin/.cookie`" >> /mnt/ssd/bitcoin/.bitcoin/restarts.log

# hold off next services for a bit
sleep 10
