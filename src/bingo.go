package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"runtime"
	"sync"
	"time"
)

var (
	WORKER_SIZE            int = 1
	TOTAL_REQUESTS             = 100
	CHUNK_SIZE_PER_REQUEST     = 10
	USE_DURATION               = false
	SITE                       = "http://localhost:8080"
	QPS                        = 5
)

type Result struct {
	Site   string
	Status int
}

func PrintResults(results <-chan Result) {

	for result := range results {
		log.Printf("Site %s finished returned with status code %d ", result.Site, result.Status)

	}
}

func makeHttpRequest(site string, result chan<- Result) (int, error) {
	resp, err := http.Get(site)
	if err != nil {
		log.Println(err.Error())
		return 500, err
	}

	defer resp.Body.Close()
	result <- Result{Status: resp.StatusCode, Site: site}

	return resp.StatusCode, nil
}

func createWorkers(ctx context.Context, results chan<- Result, wg *sync.WaitGroup, stopChan <-chan bool) <-chan bool {

	isDone := make(chan bool, WORKER_SIZE)
	wg.Add(WORKER_SIZE)
	// defer close(isDone)

	for w := 1; w <= WORKER_SIZE; w++ {

		go func(w int) {
			// wg.wait() is in a separate goroutine because
			// on the main thread it will block until all workers are done
			// and workers will only be done once work has been issued to them
			// which happens later on under a condition

			defer wg.Wait()
			worker(w, ctx, results, wg, stopChan, isDone) // put this in the background
		}(w)
	}

	return isDone

}

func listenToInterruptSignals(ctx context.Context, isDone chan<- bool) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// listen in the background
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		isDone <- true
		cancel()
	}()

	return ctx
}

func worker(wId int, ctx context.Context, results chan<- Result, wg *sync.WaitGroup, stopChan <-chan bool, isDone chan<- bool) {
	defer wg.Done()

	var throttle <-chan time.Time // declare and reassign later
	requestPerWorker := TOTAL_REQUESTS / WORKER_SIZE

	if QPS > 0 {
		// throttle = time.Tick(time.Duration(1e6/(QPS)) * time.Microsecond)
		_duration := time.Duration(time.Duration(QPS) * time.Second)
		throttle = time.Tick(_duration)

	}

	for i := 0; i < requestPerWorker; i++ {

		select {

		case <-ctx.Done():
			log.Printf("Worker %d: stopped on interrupt signal %s", wId, SITE)
			return

		case <-stopChan:
			log.Printf("Worker %d: stopped on interrupt signal %s", wId, SITE)
			return
		default:

			if QPS > 0 {
				<-throttle
			}

			log.Printf("Worker %d: %s", wId, SITE)
			status, err := makeHttpRequest(SITE, results)
			if err != nil {
				status = 500
			}

			results <- Result{Status: status, Site: SITE}
		}

	}

	isDone <- true

}

func init() {

	numCores := runtime.NumCPU()

	flag.IntVar(&TOTAL_REQUESTS, "n", TOTAL_REQUESTS, "Total number of requests to run")
	flag.IntVar(&WORKER_SIZE, "c", WORKER_SIZE, "Total number of workers to run concurrently")
	flag.BoolVar(&USE_DURATION, "z", USE_DURATION, "Enable duration based execution instead of using number of workers")
	flag.IntVar(&QPS, "q", QPS, "Rate limit, in queries per second (QPS) per worker")

	flag.Parse()

	if WORKER_SIZE > numCores {
		WORKER_SIZE = numCores
	}

	fmt.Println("tail:", flag.Args())
}

func main() {

	stopCh := make(chan bool, 1)

	ctx := context.Background()

	ctx = listenToInterruptSignals(ctx, stopCh)

	var wg sync.WaitGroup

	results := make(chan Result, TOTAL_REQUESTS)

	isDone := createWorkers(ctx, results, &wg, stopCh)

	total_count := 0

	for {

		select {
		case <-isDone:
			total_count += 1
		}

		if total_count >= WORKER_SIZE {
			close(results)
			break
		}
	}

	PrintResults(results)

}
