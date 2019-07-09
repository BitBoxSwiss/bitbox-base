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
// monitors systemd or file logs to detect potential issues and take action.
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
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//follower represents a client that follows some external resource, like a log.
type follower interface {
	// follow starts following the external resource, and continues forever, unless
	// there is an error.
	follow(chan serviceEvent, chan error)
}

// trigger is something specific that can happen for a service
type trigger int

// serviceEvent represents an actionable event from a systemd service that we are following
// e.g. that bitcoin or electrs has fully synced, or a service os not reachable
type serviceEvent struct {
	unit    string  // unit represents systemd unit name, e.g. 'bitcoind'
	trigger trigger // event could be 'fully synced', 'unit down' or 'measureUpdate'
	measure string
	value   float64
}

// logFollower follows service logs.
type logFollower struct {
	unit string            // systemd unit to follow, e.g 'bitcoind.service'
	logs chan serviceEvent // when detecting a specific trigger, a service event is sent through this chan
	errs chan error        // lines of stderr output are sent through this chan
}

// prometheusFollower follows metrics exposed by a Prometheus server.
type prometheusFollower struct {
	// unit is the systemd unit that the measure belongs to (e.g. 'bitcoind')
	unit string
	// measure is the name of the datapoint
	measure string
	// expression is the PQL expression to query for
	// if empty, measure is used
	expression string
	// server is the address of the prometheus server to connect to
	server string
	// datapoints as measures sent through this chan
	logs chan serviceEvent
	// lines of stderr output are sent through this chan
	errs chan error
}

// followers represents several follower objects.
type followers []follower

// chanStringWriter implements io.Writer and writes all contents as string into the wrapped chan.
type chanServiceEventWriter struct {
	logs chan serviceEvent
	unit string
}

// chanErrWriter implements io.Writer and writes all contents as error into the wrapped chan.
type chanErrWriter struct{ errs chan error }

// list of possible service events
const (
	fullySynced trigger = 1 + iota
	unitDown
	measureUpdate
)

var triggers = [...]string{
	"fullySynced",
	"unitDown",
	"measureUpdate",
}

func (t trigger) String() string {
	if fullySynced <= t && int(t) < len(triggers) {
		return triggers[t-1]
	}

	// log.Printf("ERR: unknown trigger '%v' encountered", t)
	return ""
	//panic(fmt.Sprintf("Invalid event: %d", int(t)))
}

// parseEvent checks a string for relevant events and potentially returns an event type
func parseEvent(p []byte, unit string) *serviceEvent {
	switch {
	// fully synched
	case strings.Contains(string(p), "finished full compaction"):
		fmt.Printf("parseEvent: 'finished full compaction'\n%q\n\n", p)
		return &serviceEvent{unit: unit, trigger: fullySynced}
	}
	return nil
}

// Write implements the io.Writer interface by sending the content as string through the wrapped channel.
func (w chanStringWriter) Write(p []byte) (int, error) {
	w.logs <- string(p)
	return len(p), nil
}

// Write implements the io.Writer interface by sending the content as error through the wrapped channel.
func (w chanErrWriter) Write(p []byte) (int, error) {
	w.errs <- fmt.Errorf(string(p))
	return len(p), nil
}

// follow indefinitely follows systemd logs for specified unit and passes
// on any output to the logs chan.
//
// If there are errors starting the journalctl command or if there is any
// output to stderr, those errors are passed on in the errs chan.
func (lf logFollower) follow() {
	args := []string{
		"--since=now",
		"--quiet",
		"--follow",
		"--unit",
		lf.unit,
	}
	cmd := exec.Command("/bin/journalctl", args...)
	fullCmd := "journalctl " + strings.Join(args, " ")
	errWriter := chanErrWriter{lf.errs}
	cmd.Stdout = chanStringWriter{lf.logs}
	cmd.Stderr = errWriter
	fmt.Printf("log follower running %q\n", fullCmd)
	if err := cmd.Run(); err != nil {
		errWriter.Write([]byte(fmt.Sprintf("failed to start cmd: %v", err)))
	}
	errWriter.Write([]byte(fmt.Sprintf("command %v unexpectedly exited", fullCmd)))
}

// follow implements follower interface by following Prometheus server and querying for values forever.
func (pf prometheusFollower) follow(events chan serviceEvent, errs chan error) {
	if len(pf.expression) == 0 {
		pf.expression = pf.measure
	}
	for {
		fmt.Printf("prometheusFollower querying for %q\n", pf.expression)
		pf.query()
		time.Sleep(5 * time.Second)
	}
}

// query queries the Prometheus server once.
func (pf prometheusFollower) query() {
	httpResp, err := http.Get(pf.server + "/api/v1/query?query=" + pf.expression)
	if err != nil {
		pf.errs <- fmt.Errorf("failed to fetch %q from prometheus server: %v", pf.expression, err)
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
		pf.errs <- fmt.Errorf("failed to read response body from prometheus request for %q: %v", pf.expression, err)
		return
	}
	fmt.Printf("xx: decoded json: %+v\n", resp)
	if resp.Status != "success" {
		pf.errs <- fmt.Errorf("prometheus request for %q returned non-success: %v", pf.expression, resp)
		return
	}
	if len(resp.Data.Result) > 1 {
		pf.errs <- fmt.Errorf("unexpectedly got %d results from prometheus request for %q: %+v", len(resp.Data.Result), pf.expression, resp)
	}
	firstResult := resp.Data.Result[0]
	if len(firstResult.Value) != 2 {
		// note: timestamp and value
		pf.errs <- fmt.Errorf("unexpectedly got %d values from prometheus request for %q: %+v", len(firstResult.Value), pf.expression, resp)
	}
	timestamp := firstResult.Value[0]
	value := firstResult.Value[1]
	fmt.Printf("xx: parsed value %+v (%T) at time %v (%T)\n", value, value, timestamp, timestamp)
	switch v := value.(type) {
	case string:
		val64, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			log.Printf("could not convert value %v of type %T into float64", value, value)
		}
		log.Printf("returned %v, %v, %v, %v", pf.unit, measureUpdate, pf.measure, val64)
		//TODO(Stadicus): replace with something like chanServiceEventWriter{events, lf.unit}
		pf.logs <- serviceEvent{unit: pf.unit, trigger: measureUpdate, measure: pf.measure, value: val64}

	default:
		pf.errs <- fmt.Errorf("unknown type of value %v (%T)", v, v)

	}
}

// restartUnit restarts a systemd unit
func restartUnit(unit string) error {
	args := []string{
		"restart",
		unit,
	}
	cmd := exec.Command("/bin/systemctl", args...)
	fullCmd := "systemctl " + strings.Join(args, " ")
	err := cmd.Run()
	if err != nil {
		log.Print(err)
		log.Printf("command '%v' unexpectedly exited", fullCmd)
	} else {
		log.Printf("restartUnit: command '%v' executed.'", fullCmd)
	}
	return err
}

func main() {
	versionNum := 0.1

	// parse command line arguments
	version := flag.Bool("version", false, "return program version")
	flag.Parse()

	fmt.Printf("bbbsupervisor version %v\n", versionNum)
	if *version {
		os.Exit(0)
	}

	// channel to process log output from systemd followers
	logs := make(chan string)
	// channel to process errors from systemd followers
	errs := make(chan error)

	// start following systemd services in separate goroutines
	fmt.Println("starting log followers..")
	followers := followers{
		logFollower{unit: "bitcoind"},
		logFollower{unit: "lightningd"},
		logFollower{unit: "electrs"},
		//prometheusFollower{unit: "lightningd", measure: "lightning_funds_output", server: "http://bob:9090"},
		prometheusFollower{unit: "lightningd", measure: "bitcoind_ibd", server: "http://bob:9090"},
	}
	for _, lf := range followers {
		go lf.follow()
	}

	for {
		select {
		// serviceEvents from systemd journal or Prometheus updates
		case message := <-events:
			fmt.Printf("follower passed on output: %v\n", message)

			switch {
			// elects: after initial sync, electrs is restarted
			case message.trigger == fullySynced:
				switch message.unit {
				case "electrs":
					fmt.Printf("Unit %v fully synced: %v\n", message.unit, message.trigger)
					restartUnit(message.unit)

				default:
					fmt.Printf("Message %v not defined for unit %v\n", message.trigger, message.unit)
				}

			case message.trigger == measureUpdate:
				fmt.Printf("Unit %v update: %v is now %v\n", message.unit, message.measure, message.value)
			}

		case err := <-errs:
			fmt.Printf("fatal: error from follower: %v\n", err)
			os.Exit(1)
			// logfile messages
			// TODO(Stadicus): tail logfiles on filesystem (if necessary)

			// Prometheus updates
			// TODO(Stadicus): recurring system metrics from Prometheus databasae
		}
	}
}
