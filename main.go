package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/sp4rd4/go-numbers-task/numbers"
)

func numbersHandler(merger *numbers.NumbersMerger, respTimeout int) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		timeout := time.After(time.Millisecond * time.Duration(respTimeout))
		if req.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Println("Obtained list of next urls:", req.URL.Query()["u"])
		merger.Merge(req.URL.Query()["u"], w, timeout)
	}
}

func main() {
	port := flag.String("port", "8000", "port for service to accept requests")
	workerCount := flag.Int("workers", 32, "`number of workers` to precess numbers")
	respTimeout := flag.Int("resp-timeout", 500, "max amount of `time in milliseconds` for service to provide answer, should be more than 50")
	reqTimeout := flag.Int("req-timeout", 450, "max amount of `time in milliseconds` to wait for answer from external service, should be more than 10")
	flag.Parse()

	mux := http.NewServeMux()
	merger := numbers.NewNumbersMerger(*workerCount, *reqTimeout)
	mux.Handle("/numbers", http.HandlerFunc(numbersHandler(merger, *respTimeout)))
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
