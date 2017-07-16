package numbers_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sp4rd4/go-numbers-task/numbers"
)

type endpoint struct {
	url          string
	numbers      []int
	responseCode int
	timeout      time.Duration
}

var mergeExamples = []struct {
	input    []endpoint
	expected []int
}{
	{
		input: []endpoint{
			{"/a", []int{2, 3, 5, 7, 11, 13}, 200, time.Millisecond * 5},
			{"/b", []int{1, 40, 32, 7, 11, 12}, 200, time.Millisecond * 5},
			{"/c", []int{19, 15, 5, 3, 1}, 200, time.Millisecond * 5},
		},
		expected: []int{1, 2, 3, 5, 7, 11, 12, 13, 15, 19, 32, 40},
	},
	{
		input: []endpoint{
			{"/a", []int{2, 3, 5, 7, 11, 13}, 503, time.Millisecond * 5},
			{"/b", []int{1, 40, 32, 7, 11, 12}, 200, time.Millisecond * 5},
			{"/c", []int{19, 15, 5, 3, 1}, 200, time.Millisecond * 5},
		},
		expected: []int{1, 3, 5, 7, 11, 12, 15, 19, 32, 40},
	},
	{
		input: []endpoint{
			{"/a", []int{2, 3, 5, 7, 11, 13}, 200, time.Millisecond * 5},
			{"/b", []int{1, 40, 32, 7, 11, 12, 1}, 200, time.Millisecond * 5},
			{"/c", []int{19, 15, 5, 3, 1}, 200, time.Millisecond * 250},
		},
		expected: []int{1, 2, 3, 5, 7, 11, 12, 13, 32, 40},
	},
	{
		input: []endpoint{
			{"/a", []int{2, 3, 5, 7, 11, 13}, 503, time.Millisecond * 5},
			{"/b", []int{1, 40, 32, 7, 11, 12, 1, 40, 22}, 200, time.Millisecond * 5},
			{"/c", []int{19, 15, 5, 3, 1}, 200, time.Millisecond * 250},
		},
		expected: []int{1, 7, 11, 12, 22, 32, 40},
	},
	{
		input: []endpoint{
			{"/a", []int{2, 3, 5, 7, 11, 13}, 503, time.Millisecond * 5},
			{"/b", []int{1, 40, 32, 7, 11, 12, 1, 40, 22}, 404, time.Millisecond * 5},
			{"/c", []int{19, 15, 5, 3, 1}, 200, time.Millisecond * 250},
		},
		expected: []int{},
	},
}

func NumberResponseStub(endpoints []endpoint) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, endpoint := range endpoints {
			if r.RequestURI == endpoint.url {
				time.Sleep(endpoint.timeout)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(endpoint.responseCode)
				json.NewEncoder(w).Encode(map[string][]int{"numbers": endpoint.numbers})
				return
			}
		}
	}))
}

func TestCorrectMerge(t *testing.T) {
	for _, example := range mergeExamples {
		buf := new(bytes.Buffer)
		logger := log.New(ioutil.Discard, "", 0)
		merger := numbers.NewNumMerger(2, 100, logger)
		defer merger.Close()
		server := NumberResponseStub(example.input)
		urls := make([]string, len(example.input))
		for i, endpoint := range example.input {
			urls[i] = server.URL + endpoint.url
		}
		merger.Merge(urls, buf, time.After(time.Millisecond*200))

		expectedBuf := new(bytes.Buffer)
		err := json.NewEncoder(expectedBuf).Encode(map[string][]int{"numbers": example.expected})
		if err != nil {
			t.Fatal(err)
		}
		if expectedBuf.String() != buf.String() {
			t.Errorf("NumMerger.Merge returned unexpected data: got %v want %v", buf.String(), expectedBuf.String())
		}
	}
}
