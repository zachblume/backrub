package main

/*
 *  DEFINITION OF PAGERANK FROM ORIGINAL PAPER
 * The Anatomy of a Large-Scale Hypertextual Web Search Engine
 * Sergey Brin and Lawrence Page
 * https://storage.googleapis.com/pub-tools-public-publication-data/pdf/334.pdf
 * We assume page A has pages T1...Tn which point to it (i.e., are citations). The parameter d
 * is a damping factor which can be set between 0 and 1. We usually set d to 0.85. There are
 * more details about d in the next section. Also C(A) is defined as the number of links going
 * out of page A. The PageRank of a page A is given as follows:
 * PR(A) = (1-d) + d (PR(T1)/C(T1) + ... + PR(Tn)/C(Tn))
 * Note that the PageRanks form a probability distribution over web pages, so the sum of all
 * web pages’ PageRanks will be one.
 * ------
 * The above definition is (infinitely) recursive.
 * For iterative computation, we will take the following steps:
 * (1) Pre-calculate the dictionary of incoming links for pages O(n) from the dict of outgoing links (database)
 * (2) Set the pagerank of every page equal to the chance of visiting a page at total random, i.e. 1/len(pages)
 *
 */
var dampingFactor float64 = 0.85

type Page struct {
	URL
	outGoingLinks
	incomingLinks
	pagerank
}

var pages []Pages

func PR(A) {
	//var solution float64 = (1-dampingFactor) + dampingFactor* PR(t[i])/C(t[i])...n

	var solution float64

	return solution
}

func main() {
	// Load database to memory

	//
}
