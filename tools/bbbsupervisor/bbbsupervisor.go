// Copyright 2019 Shift Cryptosecurity AG, Switzerland.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// BitBox Base Supervisor
// ----------------------
// Watches systemd logs (via journalctl) and queries Prometheus to detect potential issues and take action.
//
// Functionality to implement:
// * System
//   - temperature control: monitor bbbfancontrol and throttle CPU if needed
//   - disk space: monitor free space on rootfs and ssd, perform cleanup of temp & logs
//   - swap: detect issues with swap file, no memory left or "zram decompression failed", perform reboot
//
// * Middleware
//   - monitor service availability
//
// * Bitcoin Core
//   - monitor service availability
//   - perform backup tasks
//   - switch between IBD and normal operation mode (e.g. adjust dbcache)
//
// * c-lightning
//   - monitor service availability
//   - perform backup tasks (once possible)
//
// * electrs
//   - monitor service availability
//   - track initial sync and full database compaction, restart service if needed
//
// * NGINX, Grafana, ...
//   - monitor service availability
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type watcher interface {
	watch()
}

// serviceEvent represents an actionable event from a systemd service that we are following
// e.g. that bitcoin or electrs has fully synced, or a service os not reachable
type serviceEvent struct {
	unit    string  // unit represents systemd unit name, e.g. 'bitcoind'
	trigger trigger // event could be 'fully synced', 'unit down' or 'measureUpdate'
	measure string
	value   float64
}

// logWatcher watches systemd service logs.
type logWatcher struct {
	unit   string            // systemd unit to follow, e.g 'bitcoind'
	events chan serviceEvent // channel for passing service Events (e.g. a systemd log entry)
	errs   chan error        // channel for passing errors (e.g. stderr outputs)
}

// prometheusWatcher watches metrics exposed by a Prometheus server.
type prometheusWatcher struct {
	unit       string            // unit is the systemd unit that the measure belongs to (e.g. 'bitcoind')
	measure    string            // measure is the name of the datapoint
	expression string            // expression is the PQL expression to query for. If empty, measure is used.
	server     string            // server is the address of the prometheus server to query from
	interval   time.Duration     // interval query interval
	events     chan serviceEvent // channel for passing service Events (e.g. a systemd log entry)
	errs       chan error        // channel for passing errors (e.g. stderr outputs)
}

// watchers represents several follower objects.
type watchers []watcher

// errWriter implements io.Writer and writes all contents as error into the wrapped chan.
type errWriter struct{ errs chan error }

// logWriter implements io.Writer and writes all contents as string into the wrapped chan.
type logWriter struct{ logs chan string }

type eventWriter struct {
	events chan serviceEvent
	unit   string
}

// trigger is something specific that can happen for a service
type trigger int

const versionNum = 0.1
const prometheusURL = "http://localhost:9090"

const (
	triggerFullySynced = 1 + iota
	triggerUnitDown
	triggerMeasureUpdate
)

// Map of possible triggers. Mapped by their trigger to a trigger name
var triggerNames = map[trigger]string{
	triggerFullySynced:   "fullySynced",
	triggerUnitDown:      "unitDown",
	triggerMeasureUpdate: "measureUpdate",
}

// String returns a human readable value for a trigger
func (t trigger) String() string {
	if val, ok := triggerNames[t]; ok { // check if the trigger exists in the triggerNames map
		return val
	}
	return ""
}

// Write implements the io.Writer interface by sending the content as a parsed event through the event channel.
func (ew eventWriter) Write(p []byte) (int, error) {
	fmt.Printf("chanServiceEventWriter: %q\n", p)
	event := parseEvent(p, ew.unit)
	if event != nil {
		ew.events <- *event
	}
	return len(p), nil
}

// Write implements the io.Writer interface by sending the content as error through the error channel.
func (ew errWriter) Write(p []byte) (int, error) {
	ew.errs <- fmt.Errorf(string(p))
	return len(p), nil
}

// watch indefinitely watches/follows systemd logs for a specified unit.
// It passes any systemd log output on to the event channel.
// If there are errors running the journalctl command or if there is any
// output to stderr, the errors are passed on in the error channel `errs`.
func (lw logWatcher) watch() {
	systemdArgs := []string{
		"--since=now",
		"--quiet",
		"--follow",
		"--unit",
		lw.unit,
	}

	cmdAsString := "journalctl " + strings.Join(systemdArgs, " ")
	cmd := exec.Command("/bin/journalctl", systemdArgs...)

	eveWriter := eventWriter{lw.events, lw.unit}
	errWriter := errWriter{lw.errs}

	cmd.Stdout = eveWriter // stdout of journalctl is written into the events channel
	cmd.Stderr = errWriter // stderr of journalctl is written into the errs channel

	fmt.Printf("Watching journalctl for unit %s (%s) \n", lw.unit, cmdAsString)

	if err := cmd.Run(); err != nil {
		errWriter.Write([]byte(fmt.Sprintf("failed to start cmd: %v", err)))
	}
	errWriter.Write([]byte(fmt.Sprintf("command %v unexpectedly exited", cmdAsString)))
}

// watch implements watch interface by quering and watching values from a Prometheus server forever.
func (pw prometheusWatcher) watch() {
	if len(pw.expression) == 0 {
		pw.expression = pw.measure
	}
	for {
		fmt.Printf("Querying prometheus for %q\n", pw.expression)
		pw.query()
		time.Sleep(pw.interval)
	}
}

// query queries prometheus with the specified expression
func (pw prometheusWatcher) query() {

	httpResp, err := http.Get(pw.server + "/api/v1/query?query=" + pw.expression)
	if err != nil {
		pw.errs <- fmt.Errorf("failed to fetch %q from prometheus server: %v", pw.expression, err)
		return
	}
	defer httpResp.Body.Close()

	type Response struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `data:"data"`
	}

	var resp Response
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		pw.errs <- fmt.Errorf("failed to read response body from prometheus request for %q: %v", pw.expression, err)
		return
	}

	//fmt.Printf("xx: decoded json: %+v\n", resp)
	if resp.Status != "success" {
		pw.errs <- fmt.Errorf("prometheus request for %q returned non-success: %v", pw.expression, resp)
		return
	}

	if len(resp.Data.Result) > 1 {
		pw.errs <- fmt.Errorf("unexpectedly got %d results from prometheus request for %q: %+v", len(resp.Data.Result), pw.expression, resp)
	}

	firstResult := resp.Data.Result[0]
	if len(firstResult.Value) != 2 { // note: timestamp and value
		pw.errs <- fmt.Errorf("unexpectedly got %d values from prometheus request for %q: %+v", len(firstResult.Value), pw.expression, resp)
	}

	value := firstResult.Value[1]

	//timestamp := firstResult.Value[0]
	//fmt.Printf("xx: parsed value %+v (%T) at time %v (%T)\n", value, value, timestamp, timestamp)
	switch v := value.(type) {
	case string:
		val64, err := strconv.ParseFloat(value.(string), 64)

		if err != nil {
			fmt.Printf("could not convert value %v of type %T into float64", value, value)
		}

		//fmt.Printf("returned %v, %v, %v, %v \n", pw.unit, triggerMeasureUpdate, pw.measure, val64)
		pw.events <- serviceEvent{unit: pw.unit, trigger: triggerMeasureUpdate, measure: pw.measure, value: val64}
	default:
		pw.errs <- fmt.Errorf("unknown type of value %v (%T)", v, v)
	}
}

// handleFlags parses command line arguments and handles them
func handleFlags() {
	version := flag.Bool("version", false, "return program version")
	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}
}

func printVersion() {
	fmt.Printf("bbbsupervisor version %v\n", versionNum)
}

// setupWatchers sets up prometheusWatchers and logWatchers and returns them
func setupWatchers(events chan serviceEvent, errs chan error) (ws watchers) {
	return watchers{
		logWatcher{"bitcoind", events, errs},
		logWatcher{"lightningd", events, errs},
		logWatcher{"electrs", events, errs},
		prometheusWatcher{unit: "bitcoind", measure: "bitcoind_ibd", server: prometheusURL, interval: 10 * time.Second, events: events, errs: errs},
	}
}

// startWatchers starts a go routine for each watcher.
// these goroutines run indefinitely.
func startWatchers(ws watchers) {
	for _, watcher := range ws {
		go watcher.watch()
	}
}

// parseEvent checks a string for relevant events and potentially returns an event type
func parseEvent(p []byte, unit string) *serviceEvent {
	switch {
	case strings.Contains(string(p), "finished full compaction"): // fully synched electrs
		fmt.Printf("parseEvent: 'finished full compaction' %q", p)
		return &serviceEvent{unit: unit, trigger: triggerFullySynced}
	}
	return nil
}

func processEvents(events chan serviceEvent, errs chan error) {
	for {
		select {
		case err := <-errs:
			panic(fmt.Sprintf("fatal: error from watcher: %v\n", err))
		case event := <-events:
			fmt.Printf("watcher passed on output: %v\n", event)

			switch {
			case event.trigger == triggerFullySynced:
				handleElectrsFullySynced(event)
			case event.trigger == triggerMeasureUpdate:
				fmt.Printf("Unit %v update: %v is now %v\n", event.unit, event.measure, event.value)
			}
		}
	}
}

// handleElectrsFullySynced restarts electrs after the initial sync is complete
func handleElectrsFullySynced(event serviceEvent) {
	switch event.unit {
	case "electrs":
		fmt.Printf("Electrs fully synced: %v\n", event.trigger)
		restartUnit("electrs")
	default:
		fmt.Printf("Message %v not defined for unit %v\n", event.trigger, event.unit)
	}
}

// restartUnit restarts a systemd unit
func restartUnit(unit string) error {
	args := []string{"restart", unit}
	cmd := exec.Command("/bin/systemctl", args...)
	cmdAsString := "systemctl " + strings.Join(args, " ")
	err := cmd.Run()
	if err != nil {
		fmt.Errorf("Command '%v' threw an error %v", cmdAsString, err)
	} else {
		fmt.Printf("restartUnit: command '%v' executed.'", cmdAsString)
	}
	return err
}

func main() {
	handleFlags()
	printVersion()

	events := make(chan serviceEvent) // channel to process events a watcher detects
	errs := make(chan error)          // channel to process errors from watchers

	ws := setupWatchers(events, errs)
	startWatchers(ws)
	processEvents(events, errs)
}
