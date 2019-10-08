package prometheus

/* This file holds constants for the used Prometheus queries */

// BasePrometheusQuery is a string representing a Prometheus query used in the BitBox Base.
type BasePrometheusQuery string

// Queries for the Prometheus server running on the Bitbox Base
const (
	BitcoinVerificationProgress BasePrometheusQuery = "bitcoin_verification_progress"
	BitcoinBlockCount           BasePrometheusQuery = "bitcoin_blocks"
	BitcoinHeaderCount          BasePrometheusQuery = "bitcoin_headers"
	BaseSystemInfo              BasePrometheusQuery = "base_system_info"
	BaseFreeDiskspace           BasePrometheusQuery = "node_filesystem_free_bytes{fstype=\"ext4\", mountpoint=\"/mnt/ssd\"}"
	BaseTotalDiskspace          BasePrometheusQuery = "node_filesystem_size_bytes{fstype=\"ext4\", mountpoint=\"/mnt/ssd\"}"
)
