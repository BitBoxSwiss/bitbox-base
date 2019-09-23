package watcher

import "github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/trigger"

type Watcher interface {
	Watch()
}

// Event represents an event triggered by a watcher
// e.g. that bitcoin or electrs has fully synced, or a service is not reachable
type Event struct {
	Unit    string          // unit represents systemd unit name, e.g. 'bitcoind'
	Trigger trigger.Trigger // trigger could be e.g. 'triggerElectrsNoBitcoindConnectivity' or 'triggerPrometheusBitcoindIBD'
	Measure string          // measure is something that is measured by the prometheusWatcher
	Value   float64         // value is the value that has been measured
}
