package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
)

// Global constants
const MAX_WORKERS = 250

// Global vars
var depthQueue = make(chan string)
var breadthQueue = make(chan string)
var wg sync.WaitGroup
var completedURLmap = make(map[string]bool)
var mutex = &sync.Mutex{} // This is a locking mechanism to prevent simultaneous map read/write

// Data model
type webpage struct {
	URL           string
	Title         string
	OutGoingLinks []string
}

// Startup func
func main() {
	// f, _ := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// log.SetOutput(f)

	os.Remove("db.json")
	os.Remove("visited.log")

	fmt.Println("main() started")

	// Start MAX_WORKERS workers (100 by default)
	wg.Add(MAX_WORKERS)
	for i := 0; i < MAX_WORKERS-1; i++ {
		go worker()
	}

	// For now, just seed the process
	for i := 0; i < MAX_WORKERS-1; i++ {
		breadthQueue <- "https://news.ycombinator.com/show"
	}
	select {}
	wg.Wait()
	// close(breadthQueue)
	// close(depthQueue)

	fmt.Println("main() done")
}

// Safely check a map for previous visits
func haveWeAlreadyVisited(url string) bool {
	// If we use a map where each URL is 50 bytes, than we will run out of memory at ~50 million websites
	// So instead we need a external sort by chunking
	// Lets chunk by 1 million lines (~50MB)
	// But first let's do the simple version

	mutex.Lock()
	defer mutex.Unlock()
	_, visitedBefore := completedURLmap[url]
	if !visitedBefore {
		completedURLmap[url] = true
	}

	return visitedBefore
}

// Parse domain from URL
func getHost(URLtoResolve string) string {
	parsedURL, err := url.Parse(URLtoResolve)
	if err != nil {
		log.Println(err)
		return ""
	}
	return parsedURL.Host
}

func isValidURL(URLtoValidate string) bool {
	// See if the URL will parse
	_, err := url.ParseRequestURI(URLtoValidate)

	// If there is no error in parsing, and the URL begins with http...
	// then true, otherwise false (using && and operator)
	return err == nil && URLtoValidate[:4] == "http"
}

// Queue workers
func worker() {
	// Continuously grab a URL from queue
	defer wg.Done()

	for {
		select {
		case url1 := <-breadthQueue:
			process(url1)
		case url2 := <-depthQueue:
			process(url2)
		default:
			break
		}

	}

}

// Task processor
func process(url string) {

	log.Println("Completed URLS: " + strconv.Itoa(len(completedURLmap)) + " | Goroutines: " + strconv.Itoa(runtime.NumGoroutine()))

	// First, check to see if the URL is valid
	if !isValidURL(url) {
		return
	}

	// Establish HTTP connection and handle errors
	http.DefaultClient.CloseIdleConnections()
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err, "connection error")
		resp.Body.Close()
		return
	}

	// Defer closing connection to end of function scope
	resp.Body.Close()

}
