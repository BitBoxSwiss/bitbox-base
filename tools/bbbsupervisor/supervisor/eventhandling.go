package supervisor

import (
	"fmt"
	"log"
	"time"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/trigger"
	"github.com/digitalbitbox/bitbox02-api-go/api/firmware/messages"
)

/* This file includes the event parsing and handling code for the bbbsupervisor. */

// eventLoop loops indefinitely and processes incoming events
func (s *Supervisor) eventLoop() {
	for {
		s.eventHandler()
	}
}

// eventHandler handles errors and events.
// When a panic occurs the error is recovered from without stopping the eventLoop().
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
		case event.Trigger == trigger.MiddlewareBaseImageUpdateStart:
			err = s.handleBaseImageUpdateStart(event)
		case event.Trigger == trigger.MiddlewareBaseImageUpdateSuccess:
			err = s.handleBaseImageUpdateSuccess(event)
		case event.Trigger == trigger.MiddlewareBaseImageUpdateFailure:
			err = s.handleBaseImageUpdateFailure(event)
		case event.Trigger == trigger.MiddlewareRPCReboot:
			err = s.handleRPCReboot(event)
		case event.Trigger == trigger.MiddlewareRPCShutdown:
			err = s.handleRPCShutdown(event)
		default:
			panic(fmt.Errorf("trigger %d is unhandled", event.Trigger))
		}
		if err != nil {
			panic(fmt.Errorf("could not trigger %s: %s", event.Trigger.String(), err))
		}
	}
}

// handleBaseImageUpdateStart is called when the middleware logs the logtag
// `LogTagMWUpdateStart`. The UPDATE_FAILED state is deactivated and the
// DOWNLOAD_UPDATE state is activated.
func (s *Supervisor) handleBaseImageUpdateStart(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.deactivateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_UPDATE_FAILED)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.activateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.notifyMiddlewareSubsystemStateChanged()
	if err != nil {
		return fmt.Errorf("could not notify the middleware about a new systemstate for trigger %q: %w", t.String(), err)
	}

	return nil
}

// handleBaseImageUpdateSuccess is called when the Middleware logs the logtag
// `LogTagMWUpdateSuccess`. Both the UPDATE_FAILED and DOWNLOAD_UPDATE states
// are deactivated.
func (s *Supervisor) handleBaseImageUpdateSuccess(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.deactivateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_UPDATE_FAILED)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.deactivateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.notifyMiddlewareSubsystemStateChanged()
	if err != nil {
		return fmt.Errorf("could not notify the middleware about a new systemstate for trigger %q: %w", t.String(), err)
	}

	return nil
}

// handleBaseImageUpdateFailure is called when the Middleware logs the logtag
// `LogTagMWUpdateFailure`.The DOWNLOAD_UPDATE state is deactivated and the
// UPDATE_FAILED state is activated.
func (s *Supervisor) handleBaseImageUpdateFailure(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.deactivateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.activateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_UPDATE_FAILED)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.notifyMiddlewareSubsystemStateChanged()
	if err != nil {
		return fmt.Errorf("could not notify the middleware about a new systemstate for trigger %q: %w", t.String(), err)
	}

	return nil
}

// handleRPCShutdown is called when the Middleware logs the logtag
// `LogTagMWShutdown`. The SHUTDOWN state is activated. The state gets reset
// once the Supervisor restarts.
func (s *Supervisor) handleRPCShutdown(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.activateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_SHUTDOWN)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.notifyMiddlewareSubsystemStateChanged()
	if err != nil {
		return fmt.Errorf("could not notify the middleware about a new systemstate for trigger %q: %w", t.String(), err)
	}

	return nil
}

// handleRPCReboot is called when the Middleware logs the logtag
// `LogTagMWReboot`. The REBOOT state gets activated. The state gets deactivated
// once the Supervisor restarts (e.g. after the reboot).
func (s *Supervisor) handleRPCReboot(event watcher.Event) error {
	t := event.Trigger
	err := t.IsFlooding(30*time.Second, s.state.TriggerLastExecuted[t])
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.activateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_REBOOT)
	if err != nil {
		return fmt.Errorf("could not handle trigger %q: %w", t.String(), err)
	}

	err = s.notifyMiddlewareSubsystemStateChanged()
	if err != nil {
		return fmt.Errorf("could not notify the middleware about a new systemstate for trigger %q: %w", t.String(), err)
	}

	return nil
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
				return fmt.Errorf("Handling trigger %q: Initial set. Setting BBB config value to `true` failed: %s", t.String(), err)
			}
			err = s.activateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC)
			if err != nil {
				return fmt.Errorf("Handling trigger %q: Initial set. Could not activate subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
			}
			err = s.notifyMiddlewareSubsystemStateChanged()
			if err != nil {
				return fmt.Errorf("Handling trigger %q: Initial set. Could not notify middleware about activated subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
			}
		} else {
			err := s.disableBaseIBDState()
			if err != nil {
				return fmt.Errorf("Handling trigger %q: Initial state set. %s", t.String(), err.Error())
			}
			err = s.deactivateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC)
			if err != nil {
				return fmt.Errorf("Handling trigger %q: Initial set. Could not deactivate subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
			}
			err = s.notifyMiddlewareSubsystemStateChanged()
			if err != nil {
				return fmt.Errorf("Handling trigger %q: Initial set. Could not notify middleware about deactivated subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
			}
		}
		s.state.PrometheusLastStateIBD = isActive // set the initial value for the state
		return nil
	}

	if wasActive == 1 && isActive == 0 { // IBD finished
		log.Println("IBD finished: unsetting bitcoind_idb.")
		err := s.disableBaseIBDState()
		if err != nil {
			return fmt.Errorf("Handling trigger %q: %s", t.String(), err.Error())
		}
		err = s.deactivateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC)
		if err != nil {
			return fmt.Errorf("Handling trigger %q: Could not deactivate subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
		}
		err = s.notifyMiddlewareSubsystemStateChanged()
		if err != nil {
			return fmt.Errorf("Handling trigger %q: Could not notify middleware about deactivated subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
		}
		s.state.PrometheusLastStateIBD = isActive
	} else if wasActive == 0 && isActive == 1 { // IBD (re)started
		log.Println("Setting bitcoin_ibd since the IBD (re)started.")
		err := s.setBBBConfigValue("bitcoin_ibd", "true")
		if err != nil {
			return fmt.Errorf("Handling trigger %q: setting BBB config value to `true` failed: %s", t.String(), err)
		}
		err = s.activateBaseSubsystemState(messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC)
		if err != nil {
			return fmt.Errorf("Handling trigger %q: Could not activate subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
		}
		err = s.notifyMiddlewareSubsystemStateChanged()
		if err != nil {
			return fmt.Errorf("Handling trigger %q: Could not notify middleware about activated subsystem state %q: %s", t.String(), messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC.String(), err)
		}
		s.state.PrometheusLastStateIBD = isActive
	}

	return nil
}
