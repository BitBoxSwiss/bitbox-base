#!/bin/bash
#
# This script is called by the bitcoind.service AFTER starting bitcoind.
#

# We set rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie, but there seems to be
# no way to specify where to expect the bitcoin cookie for c-lightning, so let's
# create a symlink at the expected testnet location.
mkdir -p /mnt/ssd/bitcoin/.bitcoin/testnet3/
ln -fs /mnt/ssd/bitcoin/.bitcoin/.cookie /mnt/ssd/bitcoin/.bitcoin/testnet3/.cookie

# wait a few seconds before providing cookie authentication 
# as .env file for electrs and base-middleware 
sleep 10
echo -n 'RPCPASSWORD=' > /mnt/ssd/bitcoin/.bitcoin/.cookie.env
tail -c +12 /mnt/ssd/bitcoin/.bitcoin/.cookie >> /mnt/ssd/bitcoin/.bitcoin/.cookie.env

# hold off next services for a bit
sleep 10
