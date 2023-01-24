package main

import (
	"fmt"
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

	//
}

// Takes a URL from queue, and saves a list of every URL it references and the page titles to db
func scraper() {

}

// Saves row to database
func save() {

}

// Task worker
func worker(url string, linkText string, referrer string) bool {
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

	// Close connection at end of function scope
	defer resp.Body.Close()

	// Read the response body and handle errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err, "cannot get response body")
		return false
	}

	// Grab the title and save it
	titleRegEx, _ := regexp.Compile("<title[^>]*>(.*?)</title>")
	pageTitle := titleRegEx.FindAllString(string(body), 1)

	data := Webpage{title: pageTitle[0]}
	fmt.Println(data)

	// Parse for links
	linkRegEx, _ := regexp.Compile("<a[^>]+?href=\"([^\"]+?)\"[^>]*>([^<]*)</a>")
	matches := linkRegEx.FindAllString(string(body), -1)

	// Loop through links and enqueue them
	for _, match := range matches {
		enqueue(match)
	}

	markComplete(url)

	return true
}

// Put task
func enqueue(match string) {

}

func markComplete(url string) {}
