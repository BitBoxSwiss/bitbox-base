# function to set or get redis values
#
redis_set() {
    # usage: redis_set "key" "value"
    ok=$(redis-cli -h localhost -p 6379 -n 0 SET "${1}" "${2}")
    if [[ "${ok}"  != "OK" ]]; then
        echo "ERR: could not SET key ${1}"
        # exit 1
    else
        echo "INFO: set Redis key '${1}' to '${2}'"
    fi

    ok=$(redis-cli -h localhost -p 6379 -n 0 SAVE)
    if [[ "${ok}"  != "OK" ]]; then
        echo "ERR: could not SAVE to disk"
        # exit 1
    fi
}

redis_get() {
    # usage: str=$(redis_get "key")
    ok=$(redis-cli -h localhost -p 6379 -n 0 GET "${1}")
    echo "${ok}"
}
