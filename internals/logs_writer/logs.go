package logswriter

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const ACTIVE_WORKER_COUNT int = 1024

type Log struct {
	HTTPMethod   string
	ReqPath      string
	ResponseCode int
	Err          error
	CreatedAt    time.Time
}

type Logger struct {
	RootDest string
	queue    chan Log
}

type NewLoggerOptions struct {
	Destination string
	WorkerCount int
}

func NewLogger(opts *NewLoggerOptions) *Logger {
	if opts == nil {
		opts = &NewLoggerOptions{Destination: "/var/logs"}
	}
	if opts.WorkerCount < 1 {
		opts.WorkerCount = ACTIVE_WORKER_COUNT
	}
	logger := Logger{opts.Destination, make(chan Log, opts.WorkerCount)}
	for i := opts.WorkerCount; i > 0; i-- {
		go func() {
			for {
				l := <-logger.queue
				fmt.Printf("new log received in worker(%d): %v\n", i, l)

				// an example log:
				// [2024-01-01 at 13:00:00] [INFO] [POST] [/api/v1/users/all] -> the error message goes here
			}
		}()
	}
	return &logger
}

func (l *Logger) WriteHTTPLog(r *http.Request) {
	if r == nil {
		log.Println("Error: unable to process the http request struct. please provide an valid HTTP request struct")
		return
	}
	// logEntry := fmt.Sprintf("[%s] [%s] [%s] -> %s")
	log := Log{}
	l.queue <- log
}
