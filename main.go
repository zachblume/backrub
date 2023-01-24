package main

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Collect this for the whole internet and you got a graph
type Webpage struct {
	referrer string // The URL of the referring page, to build the graph
	url      string
	title    string
	linkText string
}

// Startup func
func main() {
	// Grab a task from the queue
	task := getTask()

	//
}

// Takes a URL from queue, and saves a list of every URL it references and the page titles to db
func scraper() {

}

// Saves row to database
func save() {

}

// Task worker
func worker(url string, linkText string, referrer string) {
	// Establish HTTP connection and handle errors
	resp, err := http.Get(url)
	if err != nil {
		log.Default(err, "connection error")
	}

	// What if we encounter a resource that is not a HTML page?
	notHTML := strings.Contains(resp.Header.Get("Content-Type"), "text/html")
	if notHTML {
		resp.Body.Close()
	}

	// Close connection at end of function scope
	defer resp.Body.Close()

	// Read the response body and handle errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Default(err, "cannot get response body")
	}

	// Parse for links
	r, _ := regexp.Compile("<a[^>]+?href=\"([^\"]+?)\"[^>]*>([^<]*)</a>")
	matches := r.FindAllString(String(body))

	//
}

// Put task
func enqueue() {

}
