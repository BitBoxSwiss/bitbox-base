package prometheuswatcher

import (
	"time"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/prometheus"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/trigger"
)

// PrometheusWatcher watches metrics exposed by a Prometheus server
type PrometheusWatcher struct {
	Unit       string // unit is the systemd unit that the expression belongs to (e.g. 'bitcoind')
	Expression string // expression is the PQL expression to query for.
	PClient    prometheus.Client
	Trigger    trigger.Trigger    // trigger is the trigger to fire when a expression has been read by this watcher
	Interval   time.Duration      // interval query interval
	Events     chan watcher.Event // channel for passing service Events (e.g. a systemd log entry)
	Errors     chan error         // channel for passing errors (e.g. stderr outputs)
}

// Watch implements watcher.Watch() interface by calling the watchHandler repeatedly.
func (pw PrometheusWatcher) Watch() {
	for {
		pw.watchHandler()
		<-time.After(pw.Interval)
	}
}

//by querying and watching values from a Prometheus server
func (pw PrometheusWatcher) watchHandler() {
	measuredValue, err := pw.PClient.QueryFloat64(pw.Expression)
	if err != nil {
		pw.Errors <- err
		return
	}

	pw.Events <- watcher.Event{Unit: pw.Unit, Trigger: pw.Trigger, Measure: pw.Expression, Value: measuredValue}
}
