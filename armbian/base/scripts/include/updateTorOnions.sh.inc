# function to update all .onion addresses in Redis
#
# note: bitcoind onion address is stored in Prometheus as 'bitcoin_tor_address'
#
function updateTorOnions() {
    if [[ -f /var/lib/tor/hidden_service_ssh/hostname ]]; then
        redis_set   "tor:ssh:onion" \
                    "$(cat /var/lib/tor/hidden_service_ssh/hostname)"
    fi

    if [[ -f /var/lib/tor/hidden_service_electrs/hostname ]]; then
        redis_set   "tor:electrs:onion" \
                    "$(cat /var/lib/tor/hidden_service_electrs/hostname)"
    fi

    if [[ -f /var/lib/tor/hidden_service_lightningd/hostname ]]; then
        redis_set   "tor:lightningd:onion" \
                    "$(cat /var/lib/tor/hidden_service_lightningd/hostname)"
    fi

    if [[ -f /var/lib/tor/hidden_service_bbbmiddleware/hostname ]]; then
        redis_set   "tor:bbbmiddleware:onion" \
                    "$(cat /var/lib/tor/hidden_service_bbbmiddleware/hostname)"
    fi
}
