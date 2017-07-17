package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sp4rd4/go-numbers-task/numbers"
)

// numbersHandler creates handler that launches numbers.NumMerger.Merge
// created handler also sets appropriate timeout timers, gets numbers service URLs from request URL
func numbersHandler(merger numbers.Merger, respTime int, lg *log.Logger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// set timeout
		timeout := time.After(time.Millisecond * time.Duration(respTime))
		// check if we got GET request
		if req.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// log and respond with merged numbers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		lg.Println("Obtained list of next urls:", req.URL.Query()["u"])
		merger.Merge(req.URL.Query()["u"], w, timeout)
	}
}

func main() {
	// cli flags parsing, following values define config for service
	port := flag.Int("port", 8000, "`port` for service to accept requests, value between 1 and 65535")
	workerCount := flag.Int("workers", 32, "`number` of workers to precess numbers, should be more than 0")
	respTimeout := flag.Int("resp-timeout", 500, "max amount of `time in milliseconds` for service to provide answer, should be 50 or more")
	reqTimeout := flag.Int("req-timeout", 450, "max amount of `time in milliseconds` to wait for answer from external service, should be 10 or more")
	logFile := flag.String("log", "", "`filename` of log file")
	flag.Parse()

	// check cli flags for valid values
	if *respTimeout < 50 || *reqTimeout < 10 || *workerCount < 1 || *port < 1 || *port > 65535 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// initialize logger
	var logWriter io.Writer
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			logWriter = f
			defer f.Close()
		} else {
			log.Println("Unable to create or open log file due to:", err)
		}
	}
	if logWriter == nil {
		logWriter = os.Stdout
	}
	logger := log.New(logWriter, "", log.LstdFlags)

	mux := http.NewServeMux()
	// create merger and start workers
	merger := numbers.NewNumMerger(*workerCount, *reqTimeout, logger)
	defer merger.Close()

	// launch service
	mux.Handle("/numbers", http.HandlerFunc(numbersHandler(merger, *respTimeout, logger)))
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), mux))
}
