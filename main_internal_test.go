package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testMerger struct {
	urls []string
}

func (merger *testMerger) Merge(urls []string, output io.Writer, timeout <-chan time.Time) {
	merger.urls = urls
	output.Write([]byte("success"))
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestNumbersHandlerOK(t *testing.T) {
	merger := &testMerger{}
	logger := log.New(ioutil.Discard, "", 0)
	handler := http.HandlerFunc(numbersHandler(merger, 10, logger))
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/numbers?u=1&u=2&u=3", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expectedBody := "success"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
	expectedType := "application/json"
	if rr.HeaderMap.Get("Content-Type") != expectedType {
		t.Errorf("handler returned unexpected Content-Type: got %v want %v", rr.HeaderMap.Get("Content-Type"), expectedType)
	}
	expectedURLs := []string{"1", "2", "3"}
	if !compareSlices(merger.urls, expectedURLs) {
		t.Errorf("merger got wrong urls: got %v want %v", merger.urls, expectedURLs)
	}
}
func TestNumbersHandlerMethodNotAllowed(t *testing.T) {
	merger := &testMerger{}
	logger := log.New(ioutil.Discard, "", 0)
	handler := http.HandlerFunc(numbersHandler(merger, 10, logger))
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/numbers?u=1&u=2&u=3", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
