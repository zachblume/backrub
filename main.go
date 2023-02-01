package main

import (
	"encoding/json"
	"fmt"
	"io"
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
)

// Global constants
const MAX_WORKERS = 1000

// Global vars
var queue = make(chan string)
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
	os.Remove("db.json")
	os.Remove("visited.log")

	fmt.Println("main() started")

	// Start MAX_WORKERS workers (100 by default)
	wg.Add(MAX_WORKERS)
	for i := 0; i < MAX_WORKERS; i++ {
		go worker()
	}

	// For now, just seed the process
	queue <- "https://news.ycombinator.com/show"

	wg.Wait()

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
	for url := range queue {
		// fmt.Print("FOR-WORKER-NUM-")
		// fmt.Print(workerID)
		// fmt.Print("-")
		process(url)
	}
}

// Task processor
func process(url string) {
	// Debugging
	// fmt.Println("worker started")

	// Debugging
	if rand.Intn(100) == 1 {
		log.Println("Completed URLS: " + strconv.Itoa(len(completedURLmap)) + " | Goroutines: " + strconv.Itoa(runtime.NumGoroutine()))
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

	// Establish HTTP connection and handle errors
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err, "connection error")
		return
	}

	// Defer closing connection to end of function scope
	defer resp.Body.Close()

	// What if we encounter a resource that is not a HTML page?
	isHTML := strings.Contains(resp.Header.Get("Content-Type"), "text/html")
	if !isHTML {
		markComplete(url)
		return
	}

	// Read the response body and handle errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err, "cannot get response body")
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

	for _, outGoingLinkURL := range outGoingLinks {
		// fmt.Print("out")
		// Add outgoing link to queue
		go func() {
			queue <- outGoingLinkURL
		}()

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

// Iterate through DB and calculate pageRank
func pageRank() {

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
