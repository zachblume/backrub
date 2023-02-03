package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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
	breadthQueue <- "https://www.wikipedia.org"

	wg.Wait()
	close(breadthQueue)
	close(depthQueue)

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

// Make sure we don't revisit URLs
func markComplete(URL string) {
	appendNewLineToFile("visited.log", URL)
}

// Parse relative to absolute URLs!
func parseToAbsoluteURL(URLtoResolve string, baseURL string) string {
	parsedURL, err := url.Parse(URLtoResolve)
	if err != nil {
		log.Println(err)
		return ""
	}
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Println(err)
		return ""
	}

	return base.ResolveReference(parsedURL).String()
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

// var workerIDs int

// Queue workers
func worker() {
	// workerIDs++
	// workerID := workerIDs

	// Continuously grab a URL from queue
	defer wg.Done()

	// fmt.Print("worker()")

	for {
		// fmt.Print("FOR-WORKER-NUM-")
		// fmt.Print(workerID)
		// fmt.Print("-")

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
	// Debugging
	// fmt.Println("worker started")

	// Debugging
	if rand.Intn(25) == 1 {
		log.Println("Completed URLS: " + strconv.Itoa(len(completedURLmap)) + " | Goroutines: " + strconv.Itoa(runtime.NumGoroutine()) + "| stacklen" + strconv.Itoa(len(breadthQueue)+len(depthQueue)))
		// pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

	}

	// Debug
	// fmt.Println(url)

	// First, check to see if the URL is valid
	if !isValidURL(url) {
		return
	}

	// Second, check to see if we've already visited this URL, and stop if we have?
	if haveWeAlreadyVisited(url) {
		return
	}

	// Safely establish HTTP connection without keepalive, and handle errors
	body, err := fetchWithoutKeepAlive(url)
	if err != nil {
		log.Println(err, "HTTP error, cannot make response and get body "+url)
		return
	}

	// Parse HTML for links, so we can follow them
	linkRegEx := regexp.MustCompile("<a[^>]+?href=\"[ ]*([^\"]+?)[ ]*\"[^>]*>[^<]*</a>")
	matches := linkRegEx.FindAllStringSubmatch(string(body), -1)

	outGoingLinks := []string{}

	// Loop through links and queue them to channel
	for _, match := range matches {
		// Resolve relative links to absolute link, using (current) url as base
		outGoingLinkURL := parseToAbsoluteURL(match[1], url)

		// If it's valid...
		if outGoingLinkURL != "" {
			// Add link to simplified array
			outGoingLinks = append(outGoingLinks, outGoingLinkURL)
		}
	}

	// Parse HTML for the page title and save it
	titleRegEx := regexp.MustCompile("<title[^>]*>(.*?)</title>")
	pageTitleMatches := titleRegEx.FindAllStringSubmatch(string(body), 1)
	pageTitle := ""
	if len(pageTitleMatches) > 0 {
		pageTitle = pageTitleMatches[0][1]
	}

	// Now that we have the page title, complete the database record and save it
	inspectedWebpage := webpage{
		URL:           url,
		Title:         pageTitle,
		OutGoingLinks: outGoingLinks,
	}
	saveToDB(inspectedWebpage)

	// Don't visit this URL again by accident
	markComplete(url)

	// Add each outgoing link to the correct breadth or depth queue
	for _, outGoingLinkURL := range outGoingLinks {
		// fmt.Print("out")

		wg.Add(1)
		go func(outGoingLinkURL string) {
			defer wg.Done()
			if getHost(outGoingLinkURL) == getHost(url) {
				depthQueue <- outGoingLinkURL
			} else {
				breadthQueue <- outGoingLinkURL
			}
		}(outGoingLinkURL)

	}
}

// Save completed URLs to database
func saveToDB(inspectedWebpage webpage) {
	webpageInJSON, err := json.Marshal(inspectedWebpage)
	if err != nil {
		fmt.Println(err)
		return
	}
	appendNewLineToFile("db.json", string(webpageInJSON))
}

// Iterate through DB and build a hash of words and their positions
func buildTitleIndex() {

}

// db access to similar speed as fs notes - take a look at this later: https://turriate.com/articles/making-sqlite-faster-in-go

func appendNewLineToFile(filePath string, toWrite string) {
	f, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err := f.WriteString(toWrite + "\n"); err != nil {
		log.Println(err)
	}
}

func fetchWithoutKeepAlive(URLToRequest string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URLToRequest, nil)
	if err != nil {
		// handle error
		return []byte{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		return []byte{}, err
	}
	defer resp.Body.Close()

	// What if we encounter a resource that is not a HTML page?
	isHTML := strings.Contains(resp.Header.Get("Content-Type"), "text/html")
	if !isHTML {
		markComplete(URLToRequest)
		err := errors.New("not HTML")
		return []byte{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}
