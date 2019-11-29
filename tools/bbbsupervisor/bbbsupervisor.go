package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/digitalbitbox/bitbox-base/tools/bbbsupervisor/supervisor"
)

const (
	helpText = `
	Watches systemd logs (via journalctl) and queries Prometheus to detect potential issues and take action.

	Command-line arguments:
	--help
	--redis-port   			redis port (default 6379)
	--prometheus-port   prometheus port (default 9090)
  --version
	`

	versionNum = "1.0"
)

// Command line arguments
var (
	helpArg        = flag.Bool("help", false, "show help")
	redisPort      = flag.String("redis-port", "6379", "redis server port")
	prometheusPort = flag.String("prometheus-port", "9090", "prometheus sever port")
	versionArg     = flag.Bool("version", false, "prints the version")
)

func main() {
	flag.Parse()
	handleFlags()
	s := supervisor.New(*redisPort, *prometheusPort)
	s.Start()
	s.Loop()
}

// handleFlags parses command line arguments and handles them
func handleFlags() {
	log.Printf("bbbsupervisor version %s\n", versionNum)
	if *versionArg || *helpArg {
		if *helpArg {
			fmt.Println(helpText)
		}
		os.Exit(0)
	}
}
