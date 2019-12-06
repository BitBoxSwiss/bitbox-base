// Package main provides the entry point into the middleware and accepts command line arguments.
// Once compiled, the application pipes information from bitbox-base backend services to the bitbox-wallet-app and serves as an authenticator to the bitbox-base.
package main

import (
	"flag"
	"log"
	"net/http"

	middleware "github.com/digitalbitbox/bitbox-base/middleware/src"
	"github.com/digitalbitbox/bitbox-base/middleware/src/configuration"
	"github.com/digitalbitbox/bitbox-base/middleware/src/handlers"
	"github.com/digitalbitbox/bitbox-base/middleware/src/hsm"
)

// version defines the middleware version
// The version is upgraded via semantic versioning
const version string = "0.0.1"

func main() {
	middlewarePort := flag.String("middlewareport", "8845", "Port the Middleware listens on")
	electrsRPCPort := flag.String("electrsport", "51002", "Electrs RPC port")
	dataDir := flag.String("datadir", ".base", "Directory where the Middleware persistent data, like for example the noise encryption keys, is stored")
	network := flag.String("network", "testnet", "Indicate wether Bitcoin is running on mainnet or testnet")
	bbbConfigScript := flag.String("bbbconfigscript", "/opt/shift/scripts/bbb-config.sh", "Path to the bbb-config.sh script that allows setting system configuration")
	bbbCmdScript := flag.String("bbbcmdscript", "/opt/shift/scripts/bbb-cmd.sh", "Path to the bbb-cmd.sh script that allows executing system commands")
	bbbSystemctlScript := flag.String("bbbsystemctlscript", "/opt/shift/scripts/bbb-systemctl.sh", "Path to the bbb-systemctl.sh script that allows starting and stopping services on the Base")
	prometheusURL := flag.String("prometheusurl", "http://localhost:9090", "URL of the Prometheus server")
	redisPort := flag.String("redisport", "6379", "Port of the Redis server")
	redisMock := flag.Bool("redismock", false, "Flag to use the Redis mock for development instead of connecting to a redis server")
	imageUpdateInfoURL := flag.String("updateinfourl", "https://shiftcrypto.ch/updates/base.json", "URL to query information about Base image updates from")
	notificationNamedPipePath := flag.String("notificationNamedPipePath", "/tmp/middleware-notification.pipe", "Path where the Middleware creates a named pipe to receive notifications from other processes on the BitBoxBase")
	hsmSerialPort := flag.String("hsmserialport", "/dev/ttyS0", "Serial port used to communicate with the HSM")
	flag.Parse()

	hsm := hsm.NewHSM(*hsmSerialPort)
	hsmFirmware, err := hsm.WaitForFirmware()
	if err != nil {
		log.Printf("Failed to connect to the HSM firmware: %v. Continuing without HSM.", err)
	} else {
		log.Printf("HSM serial port connected.")
	}

	config := configuration.NewConfiguration(
		configuration.Args{
			BBBCmdScript:              *bbbCmdScript,
			BBBConfigScript:           *bbbConfigScript,
			BBBSystemctlScript:        *bbbSystemctlScript,
			ElectrsRPCPort:            *electrsRPCPort,
			ImageUpdateInfoURL:        *imageUpdateInfoURL,
			MiddlewarePort:            *middlewarePort,
			MiddlewareVersion:         version,
			Network:                   *network,
			NotificationNamedPipePath: *notificationNamedPipePath,
			PrometheusURL:             *prometheusURL,
			RedisMock:                 *redisMock,
			RedisPort:                 *redisPort,
		},
	)

	logBeforeExit := func() {
		// Recover from all panics and log error before panicking again.
		if r := recover(); r != nil {
			// r is of type interface{}, just print its value
			log.Printf("%v, error detected, shutting down.", r)
			panic(r)
		}
	}
	defer logBeforeExit()

	middleware, err := middleware.NewMiddleware(config, hsmFirmware)
	if err != nil {
		log.Fatalf("error starting the middleware: %s . Is redis connected? \nIf you are running the middleware outside of the base consider setting the redis mock flag to true: '-redismock true' .", err.Error())
	}
	log.Println("--------------- Started middleware --------------")

	handlers := handlers.NewHandlers(middleware, *dataDir)
	log.Printf("Binding middleware api to port %s\n", *middlewarePort)

	if err := http.ListenAndServe(":"+*middlewarePort, handlers.Router); err != nil {
		log.Println(err.Error() + " Failed to listen for HTTP")
	}
}
