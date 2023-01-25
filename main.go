package main

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Initialize database connection
max_conns := 25
conns := make(chan *sqlite3.Stmt, max_conns)

for i := 0; i < max_conns; i++ {
  conn, err := sqlite3.Open("file:db.sqlite3?cache=shared&mode=rw")
  check(err)
  stmt, err := conn.Prepare("INSERT INTO webpages (columns) VALUES (values)")
  check(err)

  defer func() {
    stmt.Close()
    conn.Close()
  }()
  conns <- stmt
}

checkout := func() *sqlite3.Stmt {
  return <-conns
}

checkin := func(c *sqlite3.Stmt) {
  conns <- c
}

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
	worker("https://google.com", "Google", "https://www.refer.com")
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

func saveToDB(inspectedWebpage Webpage) {
	stmt := checkout()
  defer checkin(stmt)
  err := stmt.Bind(1)
  _, err = stmt.Step()
  check(err)
  var body []byte
  err = stmt.Scan(&body)
  check(err)
  err = stmt.Reset()
  check(err)
  
	return true
}
// db notes - take a look at this later: https://turriate.com/articles/making-sqlite-faster-in-go