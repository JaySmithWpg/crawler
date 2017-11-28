package parser

import (
	"bytes"
	"net/http"
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

type readCloseStub struct {
	reader *bytes.Reader
}

func (rcs *readCloseStub) Read(p []byte) (int, error) {
	i, e := rcs.reader.Read(p)
	return i, e
}

func (rcs *readCloseStub) Close() error {
	return nil
}

func stubResponse(u *url.URL, bodyText *[]byte) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       &readCloseStub{reader: bytes.NewReader(*bodyText)},
		Request:    &http.Request{URL: u},
	}
}

func TestParserHandlesAbsolutePaths(t *testing.T) {
	p := Create()

	body := []byte("<html>\n<a href=\"http://pie.com/fruit/apples.html\">test</a>\n</html>\n\n")
	rootUrl, _ := url.Parse("http://monkeyland.com/pie/food.html")

	r := stubResponse(rootUrl, &body)
	go func() {
		defer p.Close()
		p.Parse(r)
	}()

	url := <-p.Urls()
	if url.Path != "/fruit/apples.html" {
		t.Errorf("Incorrect url returned: %s", url.Path)
	}
}

func TestParserHandlesRelativePaths(t *testing.T) {
	p := Create()

	body := []byte("<a href=\"../apples.html\">test</a>\n\n")
	rootUrl, _ := url.Parse("http://monkeyland.com/pie/food.html")

	r := stubResponse(rootUrl, &body)
	go func() {
		defer p.Close()
		p.Parse(r)
	}()

	url := <-p.Urls()
	if url.Path != "/apples.html" {
		t.Errorf("Incorrect url returned: %s", url.Path)
	}
}
