package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sifatulrabbi/shadowtracker/internals/logswriter"
	"github.com/sifatulrabbi/shadowtracker/internals/tcpinterceptor"
)

func main() {
	var (
		logger  *logswriter.Logger
		action  = flag.String("action", "start", "Enter the action you want to perform")
		logsDir = flag.String("logsdir", "/var/log/shadowtracker", "Enter the logs directory path")
		target  = flag.String("target", "", "Enter the port to listen to")
		forward = flag.String("forward", "", "Enter the traffic forwarding port")
	)
	flag.Parse()
	logger = logswriter.NewLogger(&logswriter.NewLoggerOptions{Destination: *logsDir})

	if *target == "" || *forward == "" {
		log.Panicln("'-target' and '-forward' is required")
	}

	switch *action {
	case "start":
		startInterceptor(*target, *forward, logger)
		break
	default:
		log.Panicln("Please enter a valid action")
		break
	}
}

func startInterceptor(target, forward string, logger *logswriter.Logger) {
	log.Println("Starting shadowtracker...")
	if err := tcpinterceptor.NewTCPListener(target, forward, logger); err != nil {
		log.Panicln(err)
	} else {
		fmt.Println("shadowtracker stopped")
	}
}
