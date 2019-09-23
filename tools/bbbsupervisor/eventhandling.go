package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

/* This file includes the event parsing and handling code for the bbbsupervisor. */

// parseEvent checks a string for relevant events and potentially returns an event type
func parseEvent(line string, unit string) *watcherEvent {
	switch {
	// fully synched electrs
	case strings.Contains(line, "finished full compaction"):
		return &watcherEvent{unit: unit, trigger: triggerElectrsFullySynced}
	// electrs unable to connect bitcoind
	case strings.Contains(line, "WARN - reconnecting to bitcoind: no reply from daemon"):
		return &watcherEvent{unit: unit, trigger: triggerElectrsNoBitcoindConnectivity}
	// bbbmiddleware unable to connect bitcoind
	case strings.Contains(line, "GetBlockChainInfo rpc call failed"):
		return &watcherEvent{unit: unit, trigger: triggerMiddlewareNoBitcoindConnectivity}
	}
	return nil
}

// eventLoop loops indefinitely and processes incoming events
func eventLoop(events chan watcherEvent, errs chan error, pState *supervisorState) {
	for {
		eventHandler(events, errs, pState)
	}
}

// eventHandler handles errors and events.
// When a panic occours the error is recovered from without stopping the eventLoop().
func eventHandler(events chan watcherEvent, errs chan error, pState *supervisorState) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered: %s\n", r)
		}
	}()

	select {
	case err := <-errs:
		panic(fmt.Errorf("watcher error: %v", err))
	case event := <-events:
		var err error
		switch {
		case event.trigger == triggerElectrsFullySynced:
			err = handleElectrsFullySynced(event, pState)
		case event.trigger == triggerElectrsNoBitcoindConnectivity:
			err = handleElectrsNoBitcoindConnectivity(event, pState)
		case event.trigger == triggerMiddlewareNoBitcoindConnectivity:
			err = handleMiddlewareNoBitcoindConnectivity(event, pState)
		case event.trigger == triggerPrometheusBitcoindIBD:
			err = handleBitcoindIBD(event, pState)
		default:
			panic(fmt.Errorf("trigger %d is unhandled", event.trigger))
		}
		if err != nil {
			panic(fmt.Errorf("could not trigger %s: %s", triggerNames[event.trigger], err))
		}
	}
}

// isTriggerFlooding checks if a trigger is flooding
// returns an error if the trigger was executed under `minDelay` time ago
func isTriggerFlooding(minDelay time.Duration, t trigger, pState *supervisorState) error {
	if lastTimeTriggered, exists := pState.triggerLastExecuted[t]; exists {
		timeSinceLastTrigger := time.Now().Sub(time.Unix(lastTimeTriggered, 0))
		if timeSinceLastTrigger < minDelay {
			// last trigger less than `minDelay` ago
			return fmt.Errorf("trigger %s is flodding. Last executed %v (minDelay %v)", t.String(), timeSinceLastTrigger, minDelay)
		}
	}
	// no entry for that trigger exist. It can't be flooding.
	return nil
}

// handleElectrsNoBitcoindConnectivity handles the triggerElectrsNoBitcoindConnectivity
// by restarting electrs which copies the current .cookie file and reloads authorization
func handleElectrsNoBitcoindConnectivity(event watcherEvent, pState *supervisorState) error {
	err := isTriggerFlooding(30*time.Second, event.trigger, pState)
	if err != nil {
		return err
	}
	log.Printf("Handling trigger %s: restarting electrs to recreate the bitcoind `.cookie` file.\n", event.trigger.String())
	err = restartUnit("electrs")
	if err != nil {
		return fmt.Errorf("Handling trigger %s: Restarting electrs failed: %v", event.trigger.String(), err)
	}
	pState.triggerLastExecuted[event.trigger] = time.Now().Unix()
	return nil
}

// handleMiddlewareNoBitcoindConnectivity handles the triggerMiddlewareNoBitcoindConnectivity
// by restarting bbbmiddleware which copies the current .cookie file and reloads authorization
func handleMiddlewareNoBitcoindConnectivity(event watcherEvent, pState *supervisorState) error {
	err := isTriggerFlooding(30*time.Second, event.trigger, pState)
	if err != nil {
		return err
	}

	log.Printf("Handling trigger %s: restarting bbbmiddleware to recreate the bitcoind `.cookie` file.\n", event.trigger.String())
	err = restartUnit("bbbmiddleware")
	if err != nil {
		return fmt.Errorf("Handling trigger %s: Restarting bbbmiddleware failed: %v", event.trigger.String(), err)
	}
	pState.triggerLastExecuted[event.trigger] = time.Now().Unix()
	return nil
}

// handleElectrsFullySynced restarts electrs after the initial sync is complete
func handleElectrsFullySynced(event watcherEvent, pState *supervisorState) error {
	err := isTriggerFlooding(30*time.Second, event.trigger, pState)
	if err != nil {
		return err
	}
	log.Printf("Handling trigger %s: restarting Electrs.\n", event.trigger.String())
	err = restartUnit("electrs")
	if err != nil {
		return fmt.Errorf("Handling trigger %s: Restarting electrs failed: %v", event.trigger.String(), err)
	}
	pState.triggerLastExecuted[event.trigger] = time.Now().Unix()
	return nil
}

// handleBitcoindIBD handles the triggerPrometheusBitcoindIBD
// by setting (true) or unsetting (false) `bitcoin_ibd` via bbb-config.sh
func handleBitcoindIBD(event watcherEvent, pState *supervisorState) error {
	wasActive, isActive := pState.prometheusLastStateIBD, event.value
	// check if isActive is valid (either 1 or 0)
	if isActive != 1 && isActive != 0 {
		return fmt.Errorf("Handling trigger %s: isActive (%f) is invalid. Should be either 1 (IBD active) or 0 (IBD inactive)", event.trigger.String(), isActive)
	}

	if wasActive == isActive {
		return nil // no state change (do nothing)
	} else if wasActive == -1 {
		// There is no prior state. Set `bitcoin_ibd` via bbbconfig.sh to true or false  (depending on the new state) just to be sure.
		if isActive == 1 {
			err := setBBBConfigValue("bitcoin_ibd", "true")
			if err != nil {
				return fmt.Errorf("Handling trigger %s: Initial set. Setting BBB config value to `true` failed: %v", event.trigger.String(), err)
			}
		} else {
			// unset clearnet IBD redis key when the IBD is finished
			err := unsetClearnetIDB()
			if err != nil {
				return fmt.Errorf("Handling trigger %s: %s", event.trigger.String(), err)
			}
			err = setBBBConfigValue("bitcoin_ibd", "false")
			if err != nil {
				return fmt.Errorf("Handling trigger %s: Initial set. Setting BBB config value `false` failed: %v", event.trigger.String(), err)
			}
		}
		pState.prometheusLastStateIBD = isActive // set the initial value for the state
	} else if wasActive == 1 && isActive == 0 { // IBD finished
		// unset clearnet IBD redis key when the IBD is finished
		err := unsetClearnetIDB()
		if err != nil {
			return fmt.Errorf("Handling trigger %s: %s", event.trigger.String(), err)
		}
		err = setBBBConfigValue("bitcoin_ibd", "false")
		if err != nil {
			return fmt.Errorf("Handling trigger %s: setting BBB config value to `false` failed: %v", event.trigger.String(), err)
		}
		pState.prometheusLastStateIBD = isActive
	} else if wasActive == 0 && isActive == 1 { // IBD (re)started
		err := setBBBConfigValue("bitcoin_ibd", "true")
		if err != nil {
			return fmt.Errorf("Handling trigger %s: setting BBB config value to `true` failed: %v", event.trigger.String(), err)
		}
		pState.prometheusLastStateIBD = isActive
	}
	return nil
}

// unsetClearnetIDB unsets (0 - download blocks over Tor) the ibdClearnetRedisKey if set.
// The key can only be set back to 1 (download blocks over clearnet) via RPC.
func unsetClearnetIDB() (err error) {
	const ibdClearnetRedisKey string = "bitcoind:ibd-clearnet"
	isIBDClearnet, err := getRedisInt(ibdClearnetRedisKey)
	if err != nil {
		return fmt.Errorf("getting redis key %s failed: %v", ibdClearnetRedisKey, err)
	}
	if isIBDClearnet == 1 {
		log.Printf("IDB finished. Setting %s to %d.\n", ibdClearnetRedisKey, 0)
		err := setBBBConfigValue("bitcoin_ibd_clearnet", "false")
		if err != nil {
			return fmt.Errorf("disabling bitcoin_ibd_clearnet via BBB config script failed: %v", err)
		}
	}
	return
}

// restartUnit restarts a systemd unit
func restartUnit(unit string) error {
	args := []string{"restart", unit}
	cmd := exec.Command("/bin/systemctl", args...)
	cmdAsString := "systemctl " + strings.Join(args, " ")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command %s threw an error %v", cmdAsString, err)
	}
	log.Printf("restartUnit: command '%v' executed.\n", cmdAsString)
	return nil
}

// setBBBConfigValue calls `bbb-config.sh set <argument> <value>`
func setBBBConfigValue(argument string, value string) error {
	args := []string{"set", argument, value}
	executable := "/usr/local/sbin/bbb-config.sh"
	cmd := exec.Command(executable, args...)
	cmdAsString := executable + " " + strings.Join(args, " ")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command %s threw an error %v", cmdAsString, err)
	}
	log.Printf("setBBBConfigValue: command '%v' executed.\n", cmdAsString)
	return nil
}
