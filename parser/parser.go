package parser

import (
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Parser interface {
	Parse(*http.Response)
	Urls() <-chan *url.URL
	Close()
}

type parser struct {
	requests chan *http.Response
	urls     chan *url.URL
}

func Create() Parser {
	p := &parser{
		requests: make(chan *http.Response),
		urls:     make(chan *url.URL),
	}
	go p.run()
	return p
}
func (p *parser) run() {
	defer close(p.urls)
	var wg sync.WaitGroup
	for request := range p.requests {
		wg.Add(1)
		go parseBody(request, p.urls, &wg)
	}
	wg.Wait()
}

func (p *parser) Parse(r *http.Response) {
	p.requests <- r
}

func (p *parser) Urls() <-chan *url.URL {
	return p.urls
}

func (p *parser) Close() {
	close(p.requests)
}

func parseBody(request *http.Response, urls chan<- *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	//TODO: error handling
	doc, _ := html.Parse(request.Body)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "a" {
			for i := range n.Attr {
				if strings.ToLower(n.Attr[i].Key) == "href" {
					urls <- parseUrl(request.Request.URL, n.Attr[i].Val)
					break
				}
			}
		}

		//Deeper!
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
}

func parseUrl(baseUrl *url.URL, path string) *url.URL {
	//TODO: Think of the errors! Why doesn't anybody think of the errors?!
	absoluteUrl, _ := baseUrl.Parse(path)
	return absoluteUrl
}
