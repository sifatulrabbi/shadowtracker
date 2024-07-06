package logswriter

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const ACTIVE_WORKER_COUNT int = 1000

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
	mu       *sync.Mutex
	logFile  *os.File
}

type NewLoggerOptions struct {
	Destination string
	WorkerCount int
}

func NewLogger(opts *NewLoggerOptions) *Logger {
	if opts == nil {
		opts = &NewLoggerOptions{Destination: "/var/log/shadowtracker"}
	}
	if opts.WorkerCount < 1 {
		opts.WorkerCount = ACTIVE_WORKER_COUNT
	}
	logger := Logger{opts.Destination, make(chan Log, opts.WorkerCount), &sync.Mutex{}, nil}

	// verify the existence of the root destination, if does not exist attempt to create the dir
	if _, err := os.ReadDir(logger.RootDest); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(logger.RootDest, 0755); err != nil {
				if strings.Contains(err.Error(), "permission denied") {
					log.Panicf("Unable to use '%s' as root destination because of: %s. Please run shadowtracker with 'sudo'\n", logger.RootDest, err)
				} else {
					log.Panicf("Unable to use '%s' as root destination because of: %s\n", logger.RootDest, err)
				}
			}
		} else {
			log.Panicf("Unable to use '%s' as root destination because of: %s\n", logger.RootDest, err)
		}
	}

	for i := opts.WorkerCount; i > 0; i-- {
		go func() {
			for {
				l := <-logger.queue
				logger.writeToLogFile(&l)
			}
		}()
	}
	return &logger
}

func (l *Logger) WriteHTTPLog(log *Log, startTime time.Time) {
	log.TimeTaken = time.Since(startTime)
	l.queue <- *log
}

func (l *Logger) writeToLogFile(localLog *Log) {
	l.mu.Lock()
	defer l.mu.Unlock()

	filename := path.Join(l.RootDest, fmt.Sprintf("%s.log", localLog.CreatedAt.Format("2006-01-02")))
	if l.logFile == nil || l.logFile.Name() != filename {
		if f, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644); err != nil {
			log.Panicf("unable to open/create '%s' because of: %s\n", filename, err)
		} else {
			stat, _ := f.Stat()
			if stat.Size() < 1 {
				f.WriteString("[timestamp]  [client ip:server ip]  [uri]  [status code]  [bytes send]  [bytes received]  [time taken]  [http version]\n")
			}
			l.logFile = f
		}
	}
	if err := l.logFile.Sync(); err != nil {
		log.Fatalln("unable to sync the log file", err)
	}

	// [timestamp] [client_ip:server_ip] [uri] [status code] [bytes send] [bytes received] [time taken] [version]
	logLine := fmt.Sprintf("[%s]  [%s:%s]  [%s]  [%d]  [%d]  [%d]  [%s]  [%s]\n",
		localLog.CreatedAt.Format(time.RFC3339), localLog.ClientIP, localLog.ServerIP,
		localLog.ReqPath, localLog.ResponseCode, localLog.ResponseSize, localLog.BodySize,
		localLog.TimeTaken.String(), localLog.HTTPVersion)
	if _, err := l.logFile.WriteString(logLine); err != nil {
		if os.IsNotExist(err) {
			l.writeToLogFile(localLog)
		} else {
			log.Printf("Unable to write the log to '%s' because of: %s", filename, err)
		}
	}
}
