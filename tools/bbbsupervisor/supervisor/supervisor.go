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

// BitBoxBase Supervisor
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

package supervisor

import (
	"log"
	"time"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/prometheus"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/redis"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/logwatcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/prometheuswatcher"
	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/watcher/trigger"
)

// supervisorState implements a current state for the supervisor.
// the state values are filled over time
type supervisorState struct {
	TriggerLastExecuted    map[trigger.Trigger]int64 // implements a state (timestamps) when a trigger was fired (to mitigate trigger flooding)
	PrometheusLastStateIBD float64                   // implements a state for the last `bitcoin_ibd` measurement value (to detect switches ibd <-> no-ibd)
}

type Supervisor struct {
	state      supervisorState
	redis      redis.Client
	prometheus prometheus.Client
	events     chan watcher.Event
	errors     chan error
	watchers   []watcher.Watcher
}

// setupWatchers sets up prometheusWatchers and logWatchers and returns them
func (s *Supervisor) setupWatchers() {
	s.watchers = []watcher.Watcher{
		logwatcher.LogWatcher{Unit: "bitcoind", Events: s.events, Errors: s.errors},
		logwatcher.LogWatcher{Unit: "lightningd", Events: s.events, Errors: s.errors},
		logwatcher.LogWatcher{Unit: "electrs", Events: s.events, Errors: s.errors},
		logwatcher.LogWatcher{Unit: "bbbmiddleware", Events: s.events, Errors: s.errors},
		prometheuswatcher.PrometheusWatcher{Unit: "bitcoind", PClient: s.prometheus, Expression: "bitcoin_ibd", Interval: 10 * time.Second, Trigger: trigger.PrometheusBitcoindIBD, Events: s.events, Errors: s.errors},
	}
}

// startWatchers starts a go routine for each watcher.
// these goroutines run indefinitely.
func (s *Supervisor) startWatchers() {
	for _, w := range s.watchers {
		go w.Watch()
	}
}

func New(redisPort string, prometheusPort string) *Supervisor {
	s := Supervisor{
		state: supervisorState{
			TriggerLastExecuted:    make(map[trigger.Trigger]int64, 0),
			PrometheusLastStateIBD: -1,
		},
		redis:      redis.NewClient(redisPort),
		prometheus: prometheus.NewClient(prometheusPort),
		events:     make(chan watcher.Event), // channel to process events a watcher detects
		errors:     make(chan error),         // channel to process errors from watchers
	}

	return &s
}

func (s *Supervisor) Start() {
	log.Println("starting bbbsupervisor")
	s.setupWatchers()
	s.startWatchers()
}

func (s *Supervisor) Loop() {
	log.Println("starting supervisor event loop")
	s.eventLoop()
}
