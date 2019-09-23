package trigger

import (
	"fmt"
	"time"
)

// Trigger is dispached by a watcher when something happens
type Trigger int

// String returns a human readable value for a trigger
func (t Trigger) String() string {
	if val, ok := triggerNames[t]; ok { // check if the trigger exists in the triggerNames map
		return val
	}
	return ""
}

const (
	ElectrsFullySynced = 1 + iota
	ElectrsNoBitcoindConnectivity
	MiddlewareNoBitcoindConnectivity
	PrometheusBitcoindIBD
)

// Map of possible triggers. Mapped by their trigger to a trigger name
var triggerNames = map[Trigger]string{
	ElectrsFullySynced:               "electrsFullySynced",
	ElectrsNoBitcoindConnectivity:    "electrsNoBitcoindConnectivity",
	MiddlewareNoBitcoindConnectivity: "triggerMiddlewareNoBitcoindConnectivity",
	PrometheusBitcoindIBD:            "prometheusBitcoindIBD",
}

// IsFlooding checks if a trigger is flooding
// returns an error if the trigger was executed under `minDelay` time ago
func (t *Trigger) IsFlooding(minDelay time.Duration, lastTimeTriggered int64) error {
	timeSinceLastTrigger := time.Now().Sub(time.Unix(lastTimeTriggered, 0))
	if timeSinceLastTrigger < minDelay {
		// last trigger less than `minDelay` ago
		return fmt.Errorf("trigger %s is flodding. Last executed %v ago (minDelay %v)", t.String(), timeSinceLastTrigger, minDelay)
	}
	return nil
}
