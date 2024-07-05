package main

import (
	"time"

	logswriter "github.com/sifatulrabbi/shadowtracker/internals/logs_writer"
)

func main() {
	logger := logswriter.NewLogger("./")
	for i := 0; i < 100; i++ {
		for i := 0; i < 100; i++ {
			logger.WriteHTTPLog()
		}
		time.Sleep(time.Second * 1)
	}
}
