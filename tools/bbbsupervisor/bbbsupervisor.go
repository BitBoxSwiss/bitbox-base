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
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type (
	// follower represents a client that follows some external resource, like a log.
	follower interface {
		// follow starts following the external resource, and continues forever, unless
		// there is an error.
		follow()
	}
	// logFollower follows service logs.
	logFollower struct {
		unit string // systemd unit to follow, e.g 'bitcoind.service'
		// TODO(hkjn): change to more structured type for the chan below for relevant log events the supervisor cares about
		logs chan string // lines of log output are sent through this chan
		errs chan error  // lines of stderr output are sent through this chan
	}

	// prometheusFollower follows metrics exposed by a Prometheus server.
	prometheusFollower struct {
		// api is the prometheus API to connect to
		api prometheus.API
		// expression is the PQL expression to query for
		expression string
		// lines of log output are sent through this chan
		logs chan string
		// lines of stderr output are sent through this chan
		errs chan error
	}
	// followers represents several follower objects.
	followers []follower
	// chanStringWriter implements io.Writer and writes all contents as string into the wrapped chan.
	chanStringWriter struct{ logs chan string }
	// chanErrWriter implements io.Writer and writes all contents as error into the wrapped chan.
	chanErrWriter struct{ errs chan error }
)

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
func (pl prometheusFollower) follow() {
	for {
		fmt.Printf("prometheusFollower querying for %q\n", pl.expression)
		pl.query()
		time.Sleep(30000 * time.Millisecond)
	}
}

// query queries the Prometheus server once.
func (pl prometheusFollower) query() {
	ctx := context.Background()
	value, err := pl.api.Query(ctx, pl.expression, time.Now())
	if err != nil {
		pl.errs <- fmt.Errorf("failed to query prometheus server: %v\n", err)
	}

	// note: the v is of type model.Value (defined by "github.com/prometheus/common/model"),
	// we can cast it to model.Vector, which via logging by %T we can tell is the concrete
	// type
	fmt.Printf("xx: value from prometheus: %+v (%T)\n", value, value)

	switch v := value.(type) {
	case model.Vector:
		vector, ok := value.(model.Vector)
		if !ok {
			log.Fatalf("failed to assert that value is model.Vector: %v (%T)\n", value, value)
		}
		for i, v := range vector {
			// since the model.Vector type is an alias for []*model.Sample, each
			// v here is a *model.Sample
			pl.logs <- fmt.Sprintf("value of %v was %v at %v", v.Metric, v.Value, v.Timestamp)
		}
	default:
		pl.errs <- fmt.Errorf("unknown type of value: %v (%T)", v, v)
	}
}

func main() {
	versionNum := 0.1

	conf := api.Config{Address: "http://localhost:9090"}
	client, err := api.NewClient(conf)
	if err != nil {
		log.Fatalf("failed to create prometheus API client: %v\n", err)
	}
	api := prometheus.NewAPI(client)

	// Query(ctx context.Context, query string, ts time.Time) (model.Value, api.Warnings, error)

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
		logFollower{unit: "NetworkManager.service", logs: logs, errs: errs},
		logFollower{unit: "bitcoind.service", logs: logs, errs: errs},
		logFollower{unit: "electrs.service", logs: logs, errs: errs},
		prometheusFollower{expression: "lightning_funds_output", api: api, logs: logs, errs: errs},
	}
	for _, lf := range followers {
		go lf.follow()
	}

	for {
		select {
		// journald log messages
		case message := <-logs:
			fmt.Printf("follower passed on output: %q\n", message)

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
