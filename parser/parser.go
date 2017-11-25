package parser

import (
	"bytes"
	"golang.org/x/net/html"
	"net/url"
	"strings"
	"sync"
)

type Request interface {
	Url() *url.URL
	Body() *[]byte
	Headers() map[string][]string
}

func Create(requests <-chan Request) <-chan *url.URL {
	urls := make(chan *url.URL)

	go func() {
		defer close(urls)
		var wg sync.WaitGroup
		for request := range requests {
			wg.Add(1)
			go parseBody(request, urls, &wg)
		}
		wg.Wait()
	}()
	return urls
}

func parseBody(request Request, urls chan<- *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	r := bytes.NewReader(*request.Body())
	//TODO: error handling
	doc, _ := html.Parse(r)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "a" {
			for i := range n.Attr {
				if strings.ToLower(n.Attr[i].Key) == "href" {
					urls <- parseUrl(request.Url(), n.Attr[i].Val)
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
