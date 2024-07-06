package tests

import (
	"testing"
	"time"

	"github.com/sifatulrabbi/shadowtracker/internals/logswriter"
)

func TestWriteLogs(t *testing.T) {
	logger := logswriter.NewLogger(&logswriter.NewLoggerOptions{Destination: "./logs", WorkerCount: 1})
	exampleLog := logswriter.Log{
		HTTPMethod:   "POST",
		HTTPVersion:  "1.1",
		ReqPath:      "/api/v1/hello",
		ResponseCode: 200,
		BodySize:     500,
		ResponseSize: 500,
		ClientIP:     "::1",
		ServerIP:     "0.0.0.0:3000",
		TimeTaken:    time.Second * 1,
		CreatedAt:    time.Now(),
		Err:          nil,
	}
	start := time.Now()
	for i := 0; i < 10; i++ {
		for i := 0; i < 100; i++ {
			logger.WriteHTTPLog(exampleLog, start)
		}
	}
}
