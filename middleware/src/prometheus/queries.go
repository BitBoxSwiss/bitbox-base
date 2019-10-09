package prometheus

/* This file holds constants for the used Prometheus queries */

// BasePrometheusQuery is a string representing a Prometheus query used in the BitBox Base.
type BasePrometheusQuery string

// Queries for the Prometheus server running on the Bitbox Base
const (
	BitcoinVerificationProgress BasePrometheusQuery = "bitcoin_verification_progress"
	BitcoinBlockCount           BasePrometheusQuery = "bitcoin_blocks"
	BitcoinHeaderCount          BasePrometheusQuery = "bitcoin_headers"
	BitcoinPeers                BasePrometheusQuery = "bitcoin_peers"
	BitcoinIBD                  BasePrometheusQuery = "bitcoin_ibd"
	BaseSystemInfo              BasePrometheusQuery = "base_system_info"
	BaseFreeDiskspace           BasePrometheusQuery = "node_filesystem_free_bytes{fstype=\"ext4\", mountpoint=\"/mnt/ssd\"}"
	BaseTotalDiskspace          BasePrometheusQuery = "node_filesystem_size_bytes{fstype=\"ext4\", mountpoint=\"/mnt/ssd\"}"
	LightningBlocks             BasePrometheusQuery = "lightning_node_blockheight"
	ElectrsBlocks               BasePrometheusQuery = "electrs_index_height"
)
