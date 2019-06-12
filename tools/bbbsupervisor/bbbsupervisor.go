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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
)

func test() {
	jconf := sdjournal.JournalReaderConfig{
		Since: time.Duration(-15) * time.Second,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: "NetworkManager.service",
			},
		},
	}

	jr, err := sdjournal.NewJournalReader(jconf)

	if err != nil {
		panic(err)
	}

	if jr == nil {
		fmt.Println("Got a nil reader")
		return
	}

	defer jr.Close()

	jr.Follow(nil, os.Stdout)
}

func main() {

	versionNum := 0.1
	cycle := 1

	// parse command line arguments
	verbose := flag.Bool("v", false, "verbose, log internal data to stdout")
	version := flag.Bool("version", false, "return program version")
	flag.Parse()

	if *version {
		fmt.Println("bbbsupervisor version", versionNum)
		os.Exit(0)
	}

	fmt.Println("BitBox Base Supervisor, version", versionNum)

	for {
		// endless loop

		if *verbose {
			fmt.Printf("debug message: %v\n", time.Now())
		}
		fmt.Printf("x")

		test()

		time.Sleep(time.Duration(cycle) * time.Second)
	}
}
