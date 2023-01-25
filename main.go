package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Data model
type Webpage struct {
	url   string
	title string

	outGoingLinks []string
}

// Startup func
func main() {
	// Grab a task from the queue
	// ...
	// For now, just seed the process
	worker("https://google.com", "Google", "https://www.refer.com")
}

// Task worker
func worker(url string) bool {
	// First, check to see if we've already visited this URL, and stop if we have?
	if haveWeAlreadyVisited(url)
    {
        return
    }
    
    // Establish HTTP connection and handle errors
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err, "connection error")
		return false
	}

	// What if we encounter a resource that is not a HTML page?
	isHTML := strings.Contains(resp.Header.Get("Content-Type"), "text/html")
	if !isHTML {
		resp.Body.Close()
		markComplete(url)
		return false
	}

	// Defer closing connection to end of function scope
	defer resp.Body.Close()

	// Read the response body and handle errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err, "cannot get response body")
		return false
	}

	// Parse HTML for links, so we can follow them
	linkRegEx := regexp.MustCompile("<a[^>]+?href=\"([^\"]+?)\"[^>]*>[^<]*</a>")
	matches := linkRegEx.FindAllStringSubmatch(string(body), -1)

	// Loop through links and queue them to channel
	for _, match := range matches {
		queue <- url
	}

	// Parse HTML for the page title and save it
	titleRegEx := regexp.MustCompile("<title[^>]*>(.*?)</title>")
	pageTitle := titleRegEx.FindAllStringSubmatch(string(body), 1)[0][1]

	// Now that we have the page title, complete the database record and save it
	inspectedWebpage := Webpage{
		url:           url,
		title:         pageTitle,
		linkText:      linkText,
		outGoingLinks: outGoingLinks,
	}
	saveToDB(inspectedWebpage)

	// Don't visit this URL again by accident
	markComplete(url)

	return true
}

// Make sure we don't revisit URLs
func markComplete(url string) {}

// Save completed URLs to database
func saveToDB(inspectedWebpage Webpage) {
	fmt.Printf("%+v", inspectedWebpage)
}

// db notes - take a look at this later: https://turriate.com/articles/making-sqlite-faster-in-go
