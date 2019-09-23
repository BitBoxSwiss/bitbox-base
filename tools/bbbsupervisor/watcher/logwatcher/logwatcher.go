package logwatcher

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/trigger"
)

// LogWatcher watches systemd service logs.
type LogWatcher struct {
	Unit   string             // systemd unit to watch, e.g 'bitcoind'
	Events chan watcher.Event // channel for passing service Events (e.g. a systemd log entry)
	Errors chan error         // channel for passing errors (e.g. stderr outputs)
}

// ErrorWriter implements io.Writer and writes all contents as error into the wrapped channel
type ErrorWriter struct {
	errs chan error
}

// EventWriter implements io.Writer and writes all contents as events into the wrapped channel
type EventWriter struct {
	events chan watcher.Event
	unit   string
}

// Write implements the io.Writer interface by sending the content as a parsed event through the event channel.
func (ew EventWriter) Write(p []byte) (int, error) {
	// sometimes multiple log lines are read as one
	logLines := strings.Split(strings.TrimSuffix(string(p), "\n"), "\n")
	for _, line := range logLines {
		event := ew.parseEvent(line, ew.unit)
		if event != nil {
			ew.events <- *event
		}
	}

	return len(p), nil
}

// Write implements the io.Writer interface by sending the content as error through the error channel.
func (ew ErrorWriter) Write(p []byte) (int, error) {
	ew.errs <- fmt.Errorf(string(p))
	return len(p), nil
}

// Watch indefinitely watches/follows systemd logs for a specified unit.
// It passes any systemd log output on to the event channel.
// If there are errors running the journalctl command or if there is any
// output to stderr, the errors are passed on in the error channel `errs`.
func (lw LogWatcher) Watch() {
	systemdArgs := []string{
		"--since=now",
		"--quiet",
		"--follow",
		"--unit",
		lw.Unit,
	}

	cmdAsString := "journalctl " + strings.Join(systemdArgs, " ")
	cmd := exec.Command("/bin/journalctl", systemdArgs...)

	eveWriter := EventWriter{lw.Events, lw.Unit}
	errWriter := ErrorWriter{lw.Errors}

	cmd.Stdout = eveWriter // stdout of journalctl is written into the events channel
	cmd.Stderr = errWriter // stderr of journalctl is written into the errs channel

	log.Printf("Watching journalctl for unit %s (%s)\n", lw.Unit, cmdAsString)
	if err := cmd.Run(); err != nil {
		errWriter.Write([]byte(fmt.Sprintf("failed to start cmd: %v", err)))
	}
	errWriter.Write([]byte(fmt.Sprintf("command %v unexpectedly exited", cmdAsString)))
}

// parseEvent checks a string for relevant events and potentially returns an event type
func (ew EventWriter) parseEvent(line string, unit string) *watcher.Event {
	switch {
	// fully synched electrs
	case strings.Contains(line, "finished full compaction"):
		return &watcher.Event{Unit: unit, Trigger: trigger.ElectrsFullySynced}
	// electrs unable to connect bitcoind
	case strings.Contains(line, "WARN - reconnecting to bitcoind: no reply from daemon"):
		return &watcher.Event{Unit: unit, Trigger: trigger.ElectrsNoBitcoindConnectivity}
	// bbbmiddleware unable to connect bitcoind
	case strings.Contains(line, "GetBlockChainInfo rpc call failed"):
		return &watcher.Event{Unit: unit, Trigger: trigger.MiddlewareNoBitcoindConnectivity}
	}
	return nil
}
