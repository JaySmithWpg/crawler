package parser

import (
	"net/url"
	"testing"
)

type requestStub struct {
	url     *url.URL
	body    *[]byte
	headers map[string][]string
}

func (r *requestStub) Url() *url.URL {
	return r.url
}

func (r *requestStub) Body() *[]byte {
	return r.body
}

func (r *requestStub) Headers() map[string][]string {
	return r.headers
}

func TestParserHandlesAbsolutePaths(t *testing.T) {
	requests := make(chan Request)
	urls := Create(requests)
	body := []byte("<html>\n<a href=\"http://pie.com/fruit/apples.html\">test</a>\n</html>\n\n")
	rootUrl, _ := url.Parse("http://monkeyland.com/pie/food.html")

	r := &requestStub{
		url:  rootUrl,
		body: &body,
	}
	go func() {
		defer close(requests)
		requests <- r
	}()

	url := <-urls
	if url.Path != "/fruit/apples.html" {
		t.Errorf("Incorrect url returned: %s", url.Path)
	}
}

func TestParserHandlesRelativePaths(t *testing.T) {
	requests := make(chan Request)
	urls := Create(requests)
	body := []byte("<a href=\"../apples.html\">test</a>\n\n")
	rootUrl, _ := url.Parse("http://monkeyland.com/pie/food.html")

	r := &requestStub{
		url:  rootUrl,
		body: &body,
	}
	go func() {
		defer close(requests)
		requests <- r
	}()

	url := <-urls
	if url.Path != "/apples.html" {
		t.Errorf("Incorrect url returned: %s", url.Path)
	}
}
