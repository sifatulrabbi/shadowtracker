package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sifatulrabbi/shadowtracker/internals/tcpinterceptor"
)

func main() {
	action := flag.String("action", "start", "Enter the action you want to perform")
	target := flag.String("target", "", "Enter the port to listen to")
	forward := flag.String("forward", "", "Enter the traffic forwarding port")
	flag.Parse()

	fmt.Println(*action, *target, *forward)

	if *target == "" || *forward == "" {
		log.Panicln("'--target' and '--forward' is required")
	}

	switch *action {
	case "start":
		startInterceptor(*target, *forward)
		break
	default:
		log.Panicln("Please enter a valid action")
		break
	}
}

func startInterceptor(target, forward string) {
	log.Println("Starting shadowtracker...")
	if err := tcpinterceptor.NewTCPListener(target, forward); err != nil {
		log.Panicln(err)
	} else {
		fmt.Println("shadowtracker stopped")
	}
}
