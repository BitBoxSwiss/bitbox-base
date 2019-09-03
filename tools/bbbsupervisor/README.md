# bbbsupervisor

Every running system needs to be managed.
On a services level, this is the job of `systemd` which starts application services in the right order, keeps track of logs and restarts services if they crash.
On top of this default service management, an application is needed that follows custom logic for the many intricacies of the various application components.
The **BitBox Base Supervisor** `bbbsupervisor` is custom-built to monitor application logs and other system metrics, watches for very specific application messages, knows how to interpret them and can take the required action.

## Scope

The Base Supervisor combines many small monitoring tasks. Contrary to the Middleware, its task is not about relaying application communication but to keep the running system in an operational state without user interaction.

See the full documentation at <https://base.shiftcrypto.ch> for handled events.

## Installation

The [application](bbbsupervisor.go) is written in Go, compiled within Docker when using the top `make` command and the resulting binary is copied to the Armbian image during build.

## Usage

TODO(Stadicus)

## Service management

The Base Supervisor is started and managed using a simple systemd unit file:

```console
[Unit]
Description=BitBox Base Supervisor
After=local-fs.target

[Service]
Type=simple
ExecStart=/usr/local/sbin/bbbsupervisor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Program architecture

Currently, two so-called watchers are implemented. A watcher watches specific resource and triggers events. For some events, actions are defined that are being taken. These two watchers are implemented right now:

- a `logWatcher` watching systemd logs for a specific service (e.g. `bitcoind.service`) via `journalctl`
- a `prometheusWatcher` watching a specific measurement exposed via the Prometheus API

### The `logWatcher`

For each systemd service, a `logWatcher` is started in its own goroutine. It starts to `--follow` the systemd log of that unit via `journalctl`. `stdout` output is written to an `eventWriter` which parses a line (sometimes also multiple lines, known issue) into an event by performing string matching on it. Each `watcherEvent` gets assigned a trigger. The event is passed into a _event channel_ called `events`. `stderr` output is written to an `errWritter` which passes all line(s) read into an _error channel_ called `errs`.

### The `prometheusWatcher`

For each Prometheus value to watch a `prometheusWatcher` is started in its own goroutine. The `prometheusWatcher` queries a specific `measure` or `expression`. It passes a watcherEvent into the `events` channel with the `measure` and the measured `value`. The watcher then sleeps and queries again after waking back up. The query `interval` can be set for each `prometheusWatcher`.

### Event handling

Events are indefinitely read from the channels (`errs`, `events`) in the `eventLoop()` function. First errors from the `errs` channel are read (if existent) and a _panic_ is thrown (currently not _recovered_ yet). Then `events` is read and the triggers are handled in the respective handle functions. Then the event handling loop restarts.

### Currently handled event triggers (business logic)

There are currently three triggers handled: `triggerElectrsFullySynced`, `triggerElectrsNoBitcoindConnectivity` and `triggerPrometheusBitcoindIDB`. I propose to document and propose triggers (including action and rationale) in a table.

| trigger | fired when | action performed | rationale |
| ---  | --- | --- | --- |
| `triggerElectrsFullySynced` (logWatcher) | Electrs log reports `"finished full compaction"`. | Restart electrs. | Free memory after initiall full sync |
| `triggerElectrsNoBitcoindConnectivity` (logWatcher) | Electrs log reports `"WARN - reconnecting to bitcoind: no reply from daemon"` | restart electrs  | lost connection to `bitcoind` due to .cookie auth |
| `triggerMiddlewareNoBitcoindConnectivity` (logWatcher) | Middleware log reports `"GetBlockChainInfo rpc call failed"` | Restarts Middleware | lost connection to `bitcoind` due to .cookie auth |
| `triggerPrometheusBitcoindIDB` (prometheusWatcher) | read Prometheus measure `bitcoind_ibd` periodically | initial trigger or value has changed: run `bbbconfig.sh set bitcoin_idb <true|false>`; not changed: nothing | adjust dbcache and stop lightningd and electrs during initial block download |

For some triggers, a (previous) state is needed. For example `triggerPrometheusBitcoindIDB` needs the previous measurement to detect a change from _idb_ to _no-idb_. For logWatcher triggers, a flood control is implemented. I.e a trigger is only handled again after a definable `minDelay` to prevent multiple handling actions being executed at roughly the same time.

#### Adding a new trigger

To add a new trigger this procedure can be followed:

1. Add the trigger to the constants and add the name to the `map[trigger]string` called `triggerNames`.
2. When adding a new trigger for a measurement on Prometheus a new `prometheusWatcher` has to be set up in `setupWatchers()`.
3. When adding a new trigger for an existing `logWatcher` a new string matcher has to be added in parseEvent().
4. Add a switch case for the new trigger in `eventLoop()` and add a handling function.
5. Handling functions shouldn't hardcode multiple (more than two) commands. Consider writing a shell script that is being run in the handling function.

## Next steps

Next steps for the supervisor could be (in no particular order):

- If needed create a watcher that checks if a service is not running
- Properly split incoming stdout lines at a `\n`
- As `bbbsupervisor.go` grows refactor it into multiple files
- Implement proper logging
- Write unit tests for e.g. `isTriggerFlooding()`, `parseEvent()`, ...
- Read `minDelay` for the flood control, query intervals, ... from a config file (maybe a JSON file as in the other Shift projects)
- Implement proper error handling and panic recovery (bbbsupervisor should not crash on an error)
- Handle system signals stopping the execution (e.g. SIGINT, SIGQUIT, SIGTERM)
- Extend `prometheusWatcher.query()` to query for strings, ints ... (currently only `float64`)
