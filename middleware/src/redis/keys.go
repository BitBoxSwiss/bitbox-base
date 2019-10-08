package redis

/* This file includes the redis keys used on the Bitbox Base */

// BaseRedisKey is a string representing a Redis key used in the BitBox Base.
type BaseRedisKey string

// BitBox Base redis keys for configuration options.
const (
	BaseHostname      BaseRedisKey = "base:hostname"
	TorEnabled        BaseRedisKey = "tor:base:enabled"
	MiddlewareOnion   BaseRedisKey = "tor:bbbmiddleware:onion"
	BitcoindListen    BaseRedisKey = "bitcoind:listen"
	BaseVersion       BaseRedisKey = "base:version"
	BitcoindVersion   BaseRedisKey = "bitcoind:version"
	LightningdVersion BaseRedisKey = "lightningd:version"
	ElectrsVersion    BaseRedisKey = "electrs:version"
)
