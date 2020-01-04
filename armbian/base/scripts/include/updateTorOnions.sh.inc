# function to update all .onion addresses in Redis
#
# note: bitcoind onion address is stored in Prometheus as 'bitcoin_tor_address'
#
function updateTorOnions() {
    # lightningd
    if systemctl is-active -q lightningd; then
        redis_set   "tor:lightningd:onion" \
                    "$(lightning-cli --conf=/etc/lightningd/lightningd.conf getinfo | jq -r '.address[0] .address')"
    fi

    # ssh
    if [[ -f /var/lib/tor/hidden_service_ssh/hostname ]]; then
        redis_set   "tor:ssh:onion" \
                    "$(cat /var/lib/tor/hidden_service_ssh/hostname)"
    fi

    # electrs
    if [[ -f /var/lib/tor/hidden_service_electrs/hostname ]]; then
        redis_set   "tor:electrs:onion" \
                    "$(cat /var/lib/tor/hidden_service_electrs/hostname)"
    fi

    # bbbmiddleware
    if [[ -f /var/lib/tor/hidden_service_bbbmiddleware/hostname ]]; then
        redis_set   "tor:bbbmiddleware:onion" \
                    "$(cat /var/lib/tor/hidden_service_bbbmiddleware/hostname)"
    fi
}
