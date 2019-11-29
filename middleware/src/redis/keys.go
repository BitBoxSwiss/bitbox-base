package redis

/* This file includes the redis keys used on the BitBoxBase */

// BaseRedisKey is a string representing a Redis key used in the BitBoxBase.
type BaseRedisKey string

// BitBoxBase redis keys for configuration options.
const (
	BaseHostname          BaseRedisKey = "base:hostname"
	TorEnabled            BaseRedisKey = "tor:base:enabled"
	MiddlewareOnion       BaseRedisKey = "tor:bbbmiddleware:onion"
	BitcoindListen        BaseRedisKey = "bitcoind:listen"
	BaseVersion           BaseRedisKey = "base:version"
	BitcoindVersion       BaseRedisKey = "bitcoind:version"
	LightningdVersion     BaseRedisKey = "lightningd:version"
	ElectrsVersion        BaseRedisKey = "electrs:version"
	MiddlewarePasswordSet BaseRedisKey = "middleware:passwordSetup"
	MiddlewareAuth        BaseRedisKey = "middleware:auth"
	BaseSetupDone         BaseRedisKey = "base:setup"
	BaseSSHDPasswordLogin BaseRedisKey = "base:sshd:passwordlogin"
	BitcoindIBDClearnet   BaseRedisKey = "bitcoind:ibd-clearnet"
)
