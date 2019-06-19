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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type (
	// logFollower follows service logs.
	logFollower struct {
		unit string // systemd unit to follow, e.g 'bitcoind.service'
		// TODO(hkjn): change to more structured type for the chan below for relevant log events the supervisor cares about
		logs chan string // lines of log output are sent through this chan
		errs chan error  // lines of stderr output are sent through this chan
	}
	// logFollowers represents several logFollower objects.
	logFollowers []logFollower
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
	followers := logFollowers{
		{unit: "NetworkManager.service", logs: logs, errs: errs},
		{unit: "bitcoind.service", logs: logs, errs: errs},
		{unit: "electrs.service", logs: logs, errs: errs},
	}
	for _, lf := range followers {
		go lf.follow()
	}

	for {
		select {
		// journald log messages
		case message := <-logs:
			fmt.Printf("log follower passed on output: %q\n", message)

		case err := <-errs:
			fmt.Printf("fatal: error from log readers: %v\n", err)
			os.Exit(1)
			// logfile messages
			// TODO(Stadicus): tail logfiles on filesystem (if necessary)

			// Prometheus updates
			// TODO(Stadicus): recurring system metrics from Prometheus databasae
		}
	}
}
