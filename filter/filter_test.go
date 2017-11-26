package filter

import (
	"net/url"
	"testing"
)

func TestFilterBlocksDuplicates(t *testing.T) {
	requests := make(chan *url.URL)
	responses := Create(requests)

	go func() {
		defer close(requests)
		url1, _ := url.Parse("http://monkey.com/test1.html")
		url2, _ := url.Parse("http://monkey.com/test1.html")
		requests <- url1
		requests <- url2
	}()
	//First response should pass through as normal
	<-responses

	//Channel will close and other response should be nil,
	resp, ok := <-responses
	if ok {
		t.Errorf("Duplicate response returned: %s", resp.String())
	}
}
