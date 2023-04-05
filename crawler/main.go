package main

import (
	"fmt"
	"log"
	"net/http"

	"sync"
	"time"
)

const (
	workerSize              int           = 3
	jobSize                 int           = 20
	requestTimeOutInSeconds time.Duration = 4
)

type Site struct {
	Url string
}

type Result struct {
	Site
	Status int
}

func PrintResults(results <-chan Result) {

	for r := 1; r <= workerSize; r++ {

		// wait for results to arrive
		select {
		case result := <-results:
			log.Printf("Site %s finished returned with status code %d ", result.Url, result.Status)
		}

	}
}

func InitiateWork(jobChannel chan<- Site, urls []string) {
	for _, url := range urls {

		jobChannel <- Site{url}
	}
}

func makeHttpRequest(site Site, result chan<- Result) {
	resp, err := http.Get(site.Url)
	if err != nil {
		log.Println(err.Error())
		return
	}

	result <- Result{Status: resp.StatusCode, Site: site}

}

func createWorkers(jobs <-chan Site, results chan<- Result, wg *sync.WaitGroup) {

	wg.Add(workerSize)
	for w := 1; w <= workerSize; w++ {

		go func(w int) {
			// wg.wait() is in a separate goroutine because
			// on the main thread it will block until all workers are done
			// and workers will only be done once work has been issued to them
			// which happens later on under a condition

			defer wg.Wait()
			crawl(w, jobs, results, wg) // put this in the background
		}(w)
	}

}

/**
<-chan Send info/data FROM
chan<- Send info/data INTO

*/

func crawl(wId int, jobs <-chan Site, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	// jobs will be picked up like a queue..each worker will get a unique job though
	// they are listening to the same channels
	for site := range jobs {
		log.Printf("Worker %d: %s", wId, site)
		makeHttpRequest(site, results)
	}
}

func main() {

	fmt.Println("worker pools in Go")

	var wg sync.WaitGroup

	jobs := make(chan Site, jobSize)
	results := make(chan Result, jobSize)

	// 1.  creates worker pool of N-size to listen for jobs
	createWorkers(jobs, results, &wg)

	// 2. Get list of sites. Also can be read from terminal, database or file system
	urls := []string{
		"https://www.ign.com",
		"https://techcrunch.com",
		"https://tutorialedge.net",
		"https://www.tesla.com",
		"http://localhost:8080",
	}

	InitiateWork(jobs, urls)

	close(jobs)

	PrintResults(results)

}
