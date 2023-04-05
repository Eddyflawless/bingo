package worker

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	err           error
	statusCode    int
	duration      time.Duration
	connDuration  time.Duration
	offset        time.Duration
	contentLength int64
}

type Report struct {
	results chan string // can be changed
	done    chan bool
}

func (r *Report) finalize(total time.Duration) {

}

type IWork interface {
	createWorkers()
	Run()
	Stop()
	Finish()
}

// extract these implementations out
var startTime = time.Now()

func now() time.Duration {

	return time.Since(startTime)
}

// pass report as callback
func runReporter(r *Report) {

	//.. rest of the implementation
	for res := range r.results {
		//..store results in a file
		fmt.Print(res)
	}
	r.done <- true
}

// end of external utils

type Work struct {
	// in seconds
	Timeout int

	// number of workers
	C int

	// Qps is the rate limit in queries per second
	QPS float64

	Output string

	initOnce sync.Once

	report *Report

	stopCh chan struct{}

	results chan *Result

	start time.Duration
}

func (h *Work) Run() {
	// h.start = now()
	h.Init()

	// create and assign reporter channel
	go func() {
		runReporter(nil) // TODO: replace with actual reference
	}()
	h.runWorkers()
	h.Finish()

}

func (h *Work) Stop() {

	// stop workers gracefully
	for i := 0; i < h.C; i++ {
		h.stopCh <- struct{}{}
	}
}

func (h *Work) Finish() {

	close(h.results)
	total := now() - h.start

	// TODO: block until report is done
	h.report.finalize(total)
}

func (h *Work) Init() {

}

func (h *Work) makeRequest(client *http.Client) {

	h.results <- &Result{
		// rest goes here
	}
}

func (h *Work) runWorkers() {

}

func (h *Work) runWorker(client *http.Client, n int) {

	var throttle <-chan time.Time // declare and reassign later

	if h.QPS > 0 {
		throttle = time.Tick(time.Duration(1e6/(h.QPS)) * time.Microsecond)
	}

	for i := 0; i < n; i++ {
		select {
		case <-h.stopCh:
			// stop even when total-queries haven't been reached
			return
		default:
			// keep going or usual flow
			if h.QPS > 0 {
				<-throttle
			}

			h.makeRequest(client)
		}
	}
}

func min(a int, b int) int {

	if a < b {
		return a
	}
	return b

}

func NewWorker() *Work {

	// setup worker here
	return &Work{}

}
