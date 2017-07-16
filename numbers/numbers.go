package numbers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type Data struct {
	Numbers []int `json:"numbers"`
}

type job struct {
	url    string
	output chan []int
}

type NumbersMerger struct {
	jobs       chan job
	httpClient *http.Client
	logger     *log.Logger
}

func StartNewNumbersMerger(workerCount int, reqTimeout int, logger *log.Logger) *NumbersMerger {
	jobs := make(chan job)
	httpClient := &http.Client{Timeout: time.Millisecond * time.Duration(reqTimeout)}
	merger := NumbersMerger{jobs, httpClient, logger}
	for i := 0; i < workerCount; i++ {
		go merger.numbersWorker()
	}
	return &merger
}

func (merger *NumbersMerger) Merge(urls []string, output io.Writer, timeout <-chan time.Time) {
	if timeout == nil {
		timeout = time.After(time.Second * 5)
	}
	workersOutput := make(chan []int)
	store := make(map[int]struct{})
	counter := len(urls)
	go func() {
		for _, url := range urls {
			merger.jobs <- job{url, workersOutput}
		}
	}()
	for {
		select {
		case numbers := <-workersOutput:
			counter--
			for _, number := range numbers {
				_, ok := store[number]
				if !ok {
					store[number] = struct{}{}
				}
			}
			if counter < 1 {
				json.NewEncoder(output).Encode(Data{storeToKeySortedSlice(store)})
				return
			}
		case <-timeout:
			json.NewEncoder(output).Encode(Data{storeToKeySortedSlice(store)})
			return
		}
	}
}

func (merger *NumbersMerger) numbersWorker() {
	for job := range merger.jobs {
		resp, err := merger.httpClient.Get(job.url)
		if err != nil {
			merger.logger.Println(job.url, "htt.Get error:", err)
			job.output <- nil
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			merger.logger.Println(job.url, "Wrong response status")
			job.output <- nil
			continue
		}
		data := Data{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			merger.logger.Println(job.url, "JSON decode error:", err)
			job.output <- nil
			continue
		}
		job.output <- data.Numbers
	}
}

func storeToKeySortedSlice(store map[int]struct{}) []int {
	slice := make([]int, 0, len(store))
	for number := range store {
		length := len(slice)
		for i := 0; i < length; i++ {
			if number < slice[i] {
				slice = append(slice, 0)
				copy(slice[i+1:], slice[i:])
				slice[i] = number
				break
			}
		}
		if len(slice) == length {
			slice = append(slice, number)
		}
	}
	return slice
}
