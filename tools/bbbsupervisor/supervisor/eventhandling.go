package supervisor

import (
	"fmt"
	"log"
	"time"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/trigger"
)

/* This file includes the event parsing and handling code for the bbbsupervisor. */

// eventLoop loops indefinitely and processes incoming events
func (s *Supervisor) eventLoop() {
	for {
		s.eventHandler()
	}
}

// eventHandler handles errors and events.
// When a panic occours the error is recovered from without stopping the eventLoop().
func (s *Supervisor) eventHandler() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered: %s\n", r)
		}
	}()

	select {
	case err := <-s.errors:
		panic(fmt.Errorf("watcher error: %v", err))
	case event := <-s.events:
		var err error
		switch {
		case event.Trigger == trigger.ElectrsFullySynced:
			err = s.handleElectrsFullySynced(event)
		case event.Trigger == trigger.ElectrsNoBitcoindConnectivity:
			err = s.handleElectrsNoBitcoindConnectivity(event)
		case event.Trigger == trigger.MiddlewareNoBitcoindConnectivity:
			err = s.handleMiddlewareNoBitcoindConnectivity(event)
		case event.Trigger == trigger.PrometheusBitcoindIBD:
			err = s.handleBitcoindIBD(event)
		default:
			panic(fmt.Errorf("trigger %d is unhandled", event.Trigger))
		}
		if err != nil {
			panic(fmt.Errorf("could not trigger %s: %s", event.Trigger.String(), err))
		}
	}
}

// handleElectrsNoBitcoindConnectivity handles the triggerElectrsNoBitcoindConnectivity
// by restarting electrs which copies the current .cookie file and reloads authorization
func (s *Supervisor) handleElectrsNoBitcoindConnectivity(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return err
	}

	log.Printf("Handling trigger %s: restarting electrs to recreate the bitcoind `.cookie` file.\n", t.String())
	err = s.restartUnit("electrs")
	if err != nil {
		return fmt.Errorf("Handling trigger %s: Restarting electrs failed: %v", t.String(), err)
	}
	s.state.TriggerLastExecuted[t] = time.Now().Unix()
	return nil
}

// handleMiddlewareNoBitcoindConnectivity handles the triggerMiddlewareNoBitcoindConnectivity
// by restarting bbbmiddleware which copies the current .cookie file and reloads authorization
func (s *Supervisor) handleMiddlewareNoBitcoindConnectivity(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return err
	}

	log.Printf("Handling trigger %s: restarting bbbmiddleware to recreate the bitcoind `.cookie` file.\n", t.String())
	err = s.restartUnit("bbbmiddleware")
	if err != nil {
		return fmt.Errorf("Handling trigger %s: Restarting bbbmiddleware failed: %v", t.String(), err)
	}
	s.state.TriggerLastExecuted[t] = time.Now().Unix()
	return nil
}

// handleElectrsFullySynced restarts electrs after the initial sync is complete
func (s *Supervisor) handleElectrsFullySynced(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return err
	}
	log.Printf("Handling trigger %s: restarting Electrs.\n", t.String())
	err = s.restartUnit("electrs")
	if err != nil {
		return fmt.Errorf("Handling trigger %s: Restarting electrs failed: %v", t.String(), err)
	}
	s.state.TriggerLastExecuted[t] = time.Now().Unix()
	return nil
}

// handleBitcoindIBD handles the triggerPrometheusBitcoindIBD
// by setting (true) or unsetting (false) `bitcoin_ibd` via bbb-config.sh
func (s *Supervisor) handleBitcoindIBD(event watcher.Event) error {
	t := event.Trigger
	wasActive, isActive := s.state.PrometheusLastStateIBD, event.Value
	// check if isActive is valid (either 1 or 0)
	if isActive != 1 && isActive != 0 {
		return fmt.Errorf("Handling trigger %s: isActive (%f) is invalid. Should be either 1 (IBD active) or 0 (IBD inactive)", t.String(), isActive)
	}

	if wasActive == isActive {
		// state did not change, do nothing
		return nil
	}

	if wasActive == -1 { // There is no prior state. Set `bitcoin_ibd` via bbbconfig.sh to true or false (depending on the new state) just to be sure.
		log.Println("Setting bitcoin_ibd since no prior state exists.")
		if isActive == 1 {
			err := s.setBBBConfigValue("bitcoin_ibd", "true")
			if err != nil {
				return fmt.Errorf("Handling trigger %s: Initial set. Setting BBB config value to `true` failed: %v", t.String(), err)
			}
		} else {
			err := s.disableBaseIBDState()
			if err != nil {
				return fmt.Errorf("Handling trigger %s: Initial state set. %s", t.String(), err.Error())
			}
		}
		s.state.PrometheusLastStateIBD = isActive // set the initial value for the state
		return nil
	}

	if wasActive == 1 && isActive == 0 { // IBD finished
		log.Println("Setting bitcoin_ibd since the IBD finished.")
		err := s.disableBaseIBDState()
		if err != nil {
			return fmt.Errorf("Handling trigger %s: %s", t.String(), err.Error())
		}
		s.state.PrometheusLastStateIBD = isActive
	} else if wasActive == 0 && isActive == 1 { // IBD (re)started
		log.Println("Setting bitcoin_ibd since the IBD (re)started.")
		err := s.setBBBConfigValue("bitcoin_ibd", "true")
		if err != nil {
			return fmt.Errorf("Handling trigger %s: setting BBB config value to `true` failed: %v", t.String(), err)
		}
		s.state.PrometheusLastStateIBD = isActive
	}

	return nil
}
