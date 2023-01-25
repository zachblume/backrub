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
	referrer string // The URL of the referring page
	url      string
	title    string
	linkText string
}

// Startup func
func main() {

	// Grab a task from the queue
	worker("https://google.com", "Google", "https://www.refer.com")

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
	titleRegEx := regexp.MustCompile("<title[^>]*>(.*?)</title>")
	pageTitle := titleRegEx.FindAllStringSubmatch(string(body), 1)[0][1]

	inspectedWebpage := Webpage{
		url:      url,
		title:    pageTitle,
		linkText: linkText,
		referrer: referrer,
	}
	saveToDB(inspectedWebpage)

	// Parse for links
	linkRegEx := regexp.MustCompile("<a[^>]+?href=\"([^\"]+?)\"[^>]*>([^<]*)</a>")
	matches := linkRegEx.FindAllStringSubmatch(string(body), -1)

	// Loop through links and enqueue them
	for _, match := range matches {
		// Prepare a task object and enqueue it
		enqueue(Webpage{
			url: match[1],
			// title:    NULL,
			linkText: match[2],
			referrer: url,
		})
	}

	markComplete(url)

	return true
}

// Put task
func enqueue(link Webpage) {
	fmt.Printf("%+v", link)
}

func markComplete(url string) {}

func saveToDB(inspectedWebpage Webpage) {
	fmt.Printf("%+v", inspectedWebpage)
}

// db notes - take a look at this later: https://turriate.com/articles/making-sqlite-faster-in-go
