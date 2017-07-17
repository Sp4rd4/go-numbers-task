package numbers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// Data respresent numbers data structure for json marshalling and unmarshalling
type Data struct {
	Numbers []int `json:"numbers"`
}

// Merger defines interface for merging
type Merger interface {
	Merge([]string, io.Writer, <-chan time.Time)
}

// NumMerger represents numbers Merger service from obtained from URLs
// NumMerger should be closed to shutdown workers listening for the jobs after it's no longer needed
type NumMerger struct {
	jobs       chan job
	httpClient *http.Client
	logger     *log.Logger
}

// job contains url to process and channel to forward results
type job struct {
	url    string
	output chan []int
}

// NewNumMerger creates new NumMerger and loads workers to process incoming URLs
// Workers are meant to be reused for different merge tasks
// workerCount defines number of workers to start
// reqTimeout defines timeout for http requests
// logger used for logging errors
func NewNumMerger(workerCount int, reqTimeout int, logger *log.Logger) *NumMerger {
	// merger config creation
	jobs := make(chan job)
	httpClient := &http.Client{Timeout: time.Millisecond * time.Duration(reqTimeout)}
	merger := NumMerger{jobs, httpClient, logger}
	// workers load
	for i := 0; i < workerCount; i++ {
		go merger.numbersWorker()
	}
	return &merger
}

// Merge starts processing URLs to get numbers and merge them into sorted numbers slice
// resulting numbers is json encoded into output io.Writer, order of URLs processing is not guaranteed
// urls contains URLs to process
// output defines writer to consume json encoded slice of sorted numbers
// timeout signals to return sorted numbers slice immediately without waiting for number from non-processed URLs
func (merger *NumMerger) Merge(urls []string, output io.Writer, timeout <-chan time.Time) {
	// create timeout if none was given
	if timeout == nil {
		timeout = time.After(time.Second * 5)
	}
	// channel to receive output from workers
	workersOutput := make(chan []int, len(urls))
	// store is map used as set-like structure thus map value is struct{} and keys are numbers we get from workers
	store := make(map[int]struct{})
	// counters store number of jobs to process
	counter := len(urls)
	// generate jobs in goroutine
	go func() {
		for _, url := range urls {
			merger.jobs <- job{url, workersOutput}
		}
	}()
	// listen for workers output immediately after jobs generation started
	for {
		select {
		// process job output
		case numbers := <-workersOutput:
			counter--
			for _, number := range numbers {
				_, ok := store[number]
				if !ok {
					store[number] = struct{}{}
				}
			}
			// return after all jobs done for this Merge
			if counter < 1 {
				json.NewEncoder(output).Encode(Data{storeToKeySortedSlice(store)})
				return
			}
		// return after timeout signal
		case <-timeout:
			json.NewEncoder(output).Encode(Data{storeToKeySortedSlice(store)})
			return
		}
	}
}

// Close shutdowns all workers started for this mergers
func (merger *NumMerger) Close() {
	close(merger.jobs)
}

// numbersWorker is worker processing job URL and returns resulting numbers through job output channel
// errors are logged into merger logger and nil is sent to output channel
func (merger *NumMerger) numbersWorker() {
	for job := range merger.jobs {
		// merger.httpClient is http.Client with request timeout set in merger initializer
		resp, err := merger.httpClient.Get(job.url)
		// check if there's any Get request errors
		if err != nil {
			merger.logger.Println(job.url, "http.Get error:", err)
			job.output <- nil
			continue
		}
		// check if response status is not Success
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			merger.logger.Println(job.url, "Wrong response status:", resp.StatusCode)
			job.output <- nil
			continue
		}
		data := Data{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		// check for json unmarshal errors
		if err != nil {
			merger.logger.Println(job.url, "JSON decode error:", err)
			job.output <- nil
			continue
		}
		job.output <- data.Numbers
	}
}

// storeToKeySortedSlice converts set-like map to sorted slice of keys
func storeToKeySortedSlice(store map[int]struct{}) []int {
	// empty zero-length slice is preinitialized with underline array capacity equal to the map size
	slice := make([]int, 0, len(store))
	for number := range store {
		if len(slice) < 1 || slice[len(slice)-1] < number {
			// append number to slice if slice is empty or last element is smaller
			slice = append(slice, number)
		} else {
			// look for position to insert via binary search
			var lo, mid int
			hi := len(slice)
			for lo < hi {
				mid = (lo + hi) / 2
				if slice[mid] > number {
					hi = mid
				} else {
					lo = mid + 1
				}
			}
			// insert at found position
			slice = append(slice, 0)
			copy(slice[lo+1:], slice[lo:])
			slice[lo] = number
		}
	}
	return slice
}
