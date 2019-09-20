# function to set or get redis values
#
redis_set() {
    # usage: redis_set "key" "value"
    ok=$(redis-cli -h localhost -p 6379 -n 0 SET "${1}" "${2}") || true
    if [[ "${ok}"  != "OK" ]]; then
        echo "ERR: Redis could not SET key ${1}"
    else
        echo "INFO: set Redis key '${1}' to '${2}'"
    fi

    ok=$(redis-cli -h localhost -p 6379 -n 0 SAVE) || true
    if [[ "${ok}"  != "OK" ]]; then
        echo "ERR: Redis could not SAVE to disk"
    fi
}

redis_get() {
    # usage: str=$(redis_get "key")
    ok=$(redis-cli -h localhost -p 6379 -n 0 GET "${1}") || true
    echo "${ok}"
}

redis_require() {
    # checks if Redis is available, or aborts
    # usage: redis_require
    ok=$(redis-cli -h localhost -p 6379 -n 0 PING) || true
    if [[ "${ok}"  != "PONG" ]]; then
        echo "ERR: Redis not available, aborting"
        echo "REDIS_REQUIRED_BUT_NOT_RUNNING" 1>&2
        exit 1
    else
        echo "INFO: Redis available"
    fi
}
