package main

// This is the whole shebang! Collect this for the whole internet and you got a graph
type Webpage struct {
	id       int32
	referrer string // The URL of the referring page, to build the graph
	url      string
	title    string
	pagerank int32
}

// Startup func
func main() {

}

// Takes a URL from queue, and saves a list of every URL it references and the page titles to db
func scraper() {

}

// Saves row to database
func save() {

}
