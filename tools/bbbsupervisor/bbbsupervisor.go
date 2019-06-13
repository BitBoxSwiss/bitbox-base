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
//

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
)

//TODO(Stadicus): create "follower" with proper methods (like newFollower, Writer) and data fields
func startFollower(service string, journaldLogMsg chan string) {
	fmt.Println(service + ": started")

	journaldLogMsg <- "channel: " + service + ": started"

	jconf := sdjournal.JournalReaderConfig{
		Since: time.Duration(-15) * time.Second,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: service,
			},
		},
	}

	jr, err := sdjournal.NewJournalReader(jconf)

	if err != nil {
		panic(err)
	}

	if jr == nil {
		fmt.Println(service + ": got a nil reader")
		return
	}

	defer jr.Close()

	// TODO(Stadicus): use custom Writer that pipes the journal entries to the `logline` channel
	jr.Follow(nil, os.Stdout)
}

// Test only, make some beeps
func test(testMsg chan string) {
	for {
		time.Sleep(time.Duration(time.Second))
		testMsg <- "beep"
	}
}

func main() {
	versionNum := 0.1

	// parse command line arguments
	version := flag.Bool("version", false, "return program version")
	flag.Parse()

	fmt.Println("bbbsupervisor version", versionNum)
	if *version {
		os.Exit(0)
	}

	// monitoring routine and channel to process input from systemd followers
	journaldLogMsg := make(chan string)

	// follower routines for systemd services
	go startFollower("NetworkManager.service", journaldLogMsg)
	go startFollower("bitcoind.service", journaldLogMsg)
	go startFollower("electrs.service", journaldLogMsg)

	// make some beeps
	testMsg := make(chan string)
	go test(testMsg)

	for {
		select {
		// journald log messages
		case message := <-journaldLogMsg:
			fmt.Println(message)

		// logfile messages
		// TODO(Stadicus): tail logfiles on filesystem (if necessary)

		// Prometheus updates
		// TODO(Stadicus): recurring system metrics from Prometheus databasae

		// test messages
		case message := <-testMsg:
			fmt.Println("Test: " + message)
		}
	}
}
