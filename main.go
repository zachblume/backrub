package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Global vars
var queue = make(chan string)
var wg sync.WaitGroup

// Data model
type webpage struct {
	URL           string
	Title         string
	OutGoingLinks []string
}

// Startup func
func main() {
	fmt.Println("main() started")

	// Start a single worker that will fan out
	wg.Add(1)   // Increment wait group
	go worker() // Start worker

	// For now, just seed the process
	queue <- "https://www.google.com"

	wg.Wait()
}

func haveWeAlreadyVisited(url string) bool {
	// If we use a map where each URL is 50 bytes, than we will run out of memory at ~50 million websites
	// So instead we need a external sort by chunking
	// Lets chunk by 1 million lines (~50MB)

	return false
}

// Parse relative to absolute URLs!
func parseToAbsoluteURL(URLtoResolve string, baseURL string) string {
	parsedURL, err := url.Parse(URLtoResolve)
	if err != nil {
		log.Println(err)
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Println(err)
		return ""
	}

	return base.ResolveReference(parsedURL).String()
}

// Task worker
func worker() {
	// Debugging
	// fmt.Println("worker started")

	// // Decrement the wait group by 1
	// defer wg.Done()

	// Grab a URL from queue
	url := <-queue

	// Debug
	// fmt.Println(url)

	// First, check to see if we've already visited this URL, and stop if we have?
	if haveWeAlreadyVisited(url) {
		return
	}

	// Establish HTTP connection and handle errors
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err, "connection error")
		return
	}

	// What if we encounter a resource that is not a HTML page?
	isHTML := strings.Contains(resp.Header.Get("Content-Type"), "text/html")
	if !isHTML {
		resp.Body.Close()
		markComplete(url)
		return
	}

	// Defer closing connection to end of function scope
	defer resp.Body.Close()

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

		if outGoingLinkURL == "" {
			continue
		}

		// Add link to simplified array
		outGoingLinks = append(outGoingLinks, outGoingLinkURL)
	}

	// Parse HTML for the page title and save it
	titleRegEx := regexp.MustCompile("<title[^>]*>(.*?)</title>")
	pageTitleMatches := titleRegEx.FindAllStringSubmatch(string(body), 1)
	pageTitle := ""
	if len(pageTitle) > 0 {
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
		// Start additional worker
		wg.Add(1)   // Increment wait group
		go worker() // Start worker

		// Add outgoing link to queue
		queue <- outGoingLinkURL
	}
}

// Make sure we don't revisit URLs
func markComplete(url string) {}

// Save completed URLs to database
func saveToDB(inspectedWebpage webpage) {
	f, err := os.OpenFile("db.json",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	toLog, err := json.Marshal(inspectedWebpage)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := f.WriteString(string(toLog) + "\n"); err != nil {
		log.Println(err)
	}
}

// Iterate through DB and calculate pageRank
func pageRank() {

}

// Iterate through DB and build a hash of words and their positions
func buildTitleIndex() {

}

// db access to similar speed as fs notes - take a look at this later: https://turriate.com/articles/making-sqlite-faster-in-go
