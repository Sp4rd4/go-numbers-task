Task assumptions
-------

* No info about required rpm and size distribution of "numbers" json responses from external services, so no specific action were taken to take care of huge "numbers" json responses or lots of external services in single request. Specifically tuned solutions usually are good only when you definitely sure that you need that and may result in more complex code.
* Task states that service should return answers in less than 500 milliseconds and at the same time empty response can be returned only when all external services return errors or timeout. I don't see a way to guarantee constant time json marshalling and only http.Server WriteTimeout can guarantee response in 500 milliseconds, so external services processing time is limited to 500 milliseconds and after 500 milliseconds timeout json marshalling will take some time. It's possible to store marshalled merged numbers and remarshall stored numbers after each new external service response, but this seems like a lot of ineffective work and processing of one big enough response with marshaling and sorting may still take more than 500 milliseconds. If we have information about external services response json sizes that assures that response json is not bigger than some value it's possible to benchmark time of sorting and marshalling biggest possible json numbers storage and subtract that time from 500 milliseconds, resulting duration should be used to set `resp-timeout` cli flag.
* Timeout for external service response is set to 450 milliseconds, value is selected arbitrary and if needed can be changed via `req-timeout` cli flag.
* External service response json should be used altogether or not at all, it shouldn't be possible that only pars of numbers are merged into resulting numbers json

Service structure
------

Two packages:
* `numbers` defines `Merge` interface, type `numbers.NumMerger` that satisfies that interface and merges external services URLs response jsons into single sorted numbers json.
* `main` launches `http.Server`, configures `numbers.NumMerger`, logger, handles `http.Request` data so that it would be consumable by `numbers.NumMerger` and sets response headers.

`numbers.NumMerger` has configurable (via `workers` cli flag) number of workers that are launched when new instance is initialized. Those workers are meant to be reused for different numbers merge request, as while goroutines are cheap they are not free therefore we don't need to create bunch of goroutines on each request.

`numbers.NumMerger.Merge(urls []string, output io.Writer, timeout <-chan time.Time)` starts processing slice of URL strings to write sorted numbers json to specified `io.Writer`. It sends URLs to workers and waits for slice `int` responses. Workers responses are saved into set-like `map` to guarantee uniqueness until there's timeout or all URLs were processed. After that set-like `map` is converted to the sorted list and js

Workers (defined as `numbersWorker()`) just load json form URL and pass extracted numbers as `int` slice (slice is essentially a small struct with pointer inside so it's okay to send it via channel) to `numbers.NumMerger.Merge`.

Possible improvements
------
Instead of processing all external services output in single goroutine it's possible to use some kind of ctrie structure for non-blocking inserts directly from workers.
