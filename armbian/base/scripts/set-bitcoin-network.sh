#!/bin/bash
set -e

# BitBox Base: set Bitcoin network
# 

function usage() {
  echo "BitBox Base: set Bitcoin network"
  echo "Usage: $0 <testnet|mainnet>"
}

if [ "$#" -ne 1 ] || ( [ "$1" != "testnet" ] && [ "$1" != "mainnet" ] ); then
  usage
  exit 1
fi

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

NETWORK="$1"

if [ "$NETWORK" == "mainnet" ]; then 
    sed -i '/CONFIGURED FOR/Ic\echo "Configured for Bitcoin MAINNET"; echo' /etc/update-motd.d/20-shift
    sed -i "/ALIAS BLOG=/Ic\alias blog='tail -f /mnt/ssd/bitcoin/.bitcoin/debug.log'" /root/.bashrc-custom
    sed -i "/ALIAS LCLI=/Ic\alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning'" /root/.bashrc-custom
    sed -i '/HIDDENSERVICEPORT 18333/Ic\HiddenServicePort 8333 127.0.0.1:8333' /etc/tor/torrc
    sed -i '/TESTNET=/Ic\#testnet=1' /etc/bitcoin/bitcoin.conf
    sed -i '/NETWORK=/Ic\network=bitcoin' /etc/lightningd/lightningd.conf
    sed -i '/BITCOIN-RPCPORT=/Ic\bitcoin-rpcport=8332' /etc/lightningd/lightningd.conf
    sed -i '/LIGHTNING-DIR=/Ic\lightning-dir=/mnt/ssd/bitcoin/.lightning' /etc/lightningd/lightningd.conf
    sed -i '/NETWORK=/Ic\NETWORK=mainnet' /etc/electrs/electrs.conf
    sed -i '/RPCPORT=/Ic\RPCPORT=8332' /etc/electrs/electrs.conf
    sed -i '/<PORT>18333/Ic\<port>8333</port>' /etc/avahi/services/bitcoind.service
else
    sed -i '/CONFIGURED FOR/Ic\echo "Configured for Bitcoin TESTNET"; echo' /etc/update-motd.d/20-shift
    sed -i "/ALIAS BLOG=/Ic\alias blog='tail -f /mnt/ssd/bitcoin/.bitcoin/testnet3/debug.log'" /root/.bashrc-custom
    sed -i "/ALIAS LCLI=/Ic\alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet'" /root/.bashrc-custom
    sed -i '/HIDDENSERVICEPORT 8333/Ic\HiddenServicePort 18333 127.0.0.1:18333' /etc/tor/torrc
    sed -i '/TESTNET=/Ic\testnet=1' /etc/bitcoin/bitcoin.conf
    sed -i '/NETWORK=/Ic\network=testnet' /etc/lightningd/lightningd.conf
    sed -i '/LIGHTNING-DIR=/Ic\lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet' /etc/lightningd/lightningd.conf
    sed -i '/BITCOIN-RPCPORT=/Ic\bitcoin-rpcport=18332' /etc/lightningd/lightningd.conf
    sed -i '/NETWORK=/Ic\NETWORK=testnet' /etc/electrs/electrs.conf
    sed -i '/RPCPORT=/Ic\RPCPORT=18332' /etc/electrs/electrs.conf
    sed -i '/<PORT>8333/Ic\<port>18333</port>' /etc/avahi/services/bitcoind.service
fi
source /root/.bashrc-custom