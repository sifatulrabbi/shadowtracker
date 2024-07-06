package logswriter

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

const ACTIVE_WORKER_COUNT int = 1024

type Log struct {
	HTTPMethod   string
	HTTPVersion  string
	ReqPath      string
	ResponseCode int
	BodySize     int
	ResponseSize int
	ClientIP     string
	ServerIP     string
	TimeTaken    time.Duration
	CreatedAt    time.Time
	Err          error
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

	// verify the existence of the root destination, if does not exist attempt to create the dir
	if _, err := os.ReadDir(logger.RootDest); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(logger.RootDest, 0755); err != nil {
				log.Panicf("Unable to use '%s' as root destination because of: %s\n", logger.RootDest, err)
			}
		} else {
			log.Panicf("Unable to use '%s' as root destination because of: %s\n", logger.RootDest, err)
		}
	}

	log.Printf("spawning %d log workers\n", opts.WorkerCount)
	for i := opts.WorkerCount; i > 0; i-- {
		go func() {
			for {
				l := <-logger.queue
				fmt.Printf("new log received in worker(%d): %v\n", i, l)
				filename := path.Join(logger.RootDest, fmt.Sprintf("%s.log", l.CreatedAt.Format(time.RFC3339)))
				f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					log.Panicf("Unable to open '%s' due to: %s", filename, err)
					continue
				}
				defer f.Close()
				// [timestamp] [client_ip:server_ip] [uri] [status code] [bytes send] [bytes received] [time taken] [version]
				logLine := fmt.Sprintf("[%s] [%s:%s] [%s] [%d] [%d] [%d] [%s] [%s]\n",
					l.CreatedAt.Format(time.RFC3339), l.ClientIP, l.ServerIP, l.ReqPath, l.ResponseCode, l.ResponseSize, l.BodySize, l.TimeTaken.String(), l.HTTPVersion)
				if _, err = f.WriteString(logLine); err != nil {
					log.Printf("Unable to write the log to '%s' due to: %s", filename, err)
				} else {
					log.Println("New log added")
				}
			}
		}()
	}
	return &logger
}

func (l *Logger) WriteHTTPLog(log Log, startTime time.Time) {
	fmt.Println("writing a new log")
	log.TimeTaken = time.Since(startTime)
	l.queue <- log
}
