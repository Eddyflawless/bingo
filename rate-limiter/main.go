package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	RATE_LIMIT_IN_SECONDS = 5
	WORK_POOL_SIZE        = 4
)

var (
	TOTAL_REQUESTS         = 100
	CHUNK_SIZE_PER_REQUEST = 10
	TARGET_URL             = "http://localhost:3000"
	USE_DURATION           = false
)

var taskQueue = make(chan int, CHUNK_SIZE_PER_REQUEST)

type Chunks struct {
	items []int
}

func createChunk() []int {
	return make([]int, CHUNK_SIZE_PER_REQUEST)
}

func CreateMockRequests(requestSize int) chan []int {

	requests := make(chan []int, requestSize/CHUNK_SIZE_PER_REQUEST)
	sub_requests := createChunk()
	for i := 1; i < requestSize; i++ {

		ptr := i % CHUNK_SIZE_PER_REQUEST
		if ptr == 0 {
			requests <- sub_requests
			sub_requests = createChunk()
		}
		sub_requests[ptr] = i
	}
	close(requests)
	return requests
}

func createLimiter() <-chan time.Time {

	return time.Tick(RATE_LIMIT_IN_SECONDS * time.Second)
}

func handleRequest(workerId int) {

	// listen to queue
	for task := range taskQueue {
		fmt.Printf("Received task %v \n ", task)
		fmt.Printf("WorkerId: %d - Processing chunk task: %d\n", workerId, task)
	}
}

func createWorkers() {

	for workerId := 0; workerId < WORK_POOL_SIZE; workerId++ {

		fmt.Printf("Creating worker %d\n", workerId)
		go handleRequest(workerId)
	}
}
func processChunkRequest(chunk []int) {

	for i := range chunk {
		fmt.Printf("Processing chunk: %d\n", chunk[i])
		// push to worker queue
		taskQueue <- chunk[i]
	}
}

func ReadMockRequests(chunks <-chan []int, limiter <-chan time.Time, notifier chan<- bool) {

	for chunk := range chunks {
		fmt.Println("Waiting for next chunk...")
		// wait till we receive a signal based on ticker time
		// limiter might not be enough
		<-limiter
		fmt.Println("chunk received", chunk, time.Now())
		processChunkRequest(chunk)

	}

	notifier <- true
	close(notifier)

}

func listenToInterruptSignals(isDone chan<- bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// listen in the background
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		isDone <- true
	}()
}

func cleanUp() {
	close(taskQueue)
}

func main() {

	isDone := make(chan bool, 1)

	listenToInterruptSignals(isDone)

	// create workers
	createWorkers()
	chunks := CreateMockRequests(TOTAL_REQUESTS)
	// create request limiter
	limiter := createLimiter()

	// run in the background
	go func() { ReadMockRequests(chunks, limiter, isDone) }()

	<-isDone

	cleanUp()
	fmt.Println("Request is done 2.")

}
