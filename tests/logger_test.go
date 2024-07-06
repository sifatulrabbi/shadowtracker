package tests

import (
	"testing"
	"time"

	"github.com/sifatulrabbi/shadowtracker/internals/logswriter"
)

func TestWriteLogs(t *testing.T) {
	logger := logswriter.NewLogger(&logswriter.NewLoggerOptions{Destination: "./logs"})
	exampleLog := logswriter.Log{}
	start := time.Now()
	for i := 0; i < 100; i++ {
		for i := 0; i < 100; i++ {
			logger.WriteHTTPLog(exampleLog, start)
		}
		time.Sleep(time.Second * 1)
	}
}
