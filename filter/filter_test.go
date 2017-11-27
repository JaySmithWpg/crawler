package filter

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func timer(signal chan bool, t time.Duration) {
	defer close(signal)
	time.Sleep(t)
}

func TestIncrementalBackoffOnError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running incremental back-off test")
	}
	filter := Create()
	defer filter.Close()

	var timeSignal chan bool
	responseUrl, _ := url.Parse("http://monkey.com/test")
	response := &http.Response{
		StatusCode: 500,
		Request:    &http.Request{URL: responseUrl},
	}

	var request *url.URL
	// the first request should pass the filter in under 10ms
	timeSignal = make(chan bool)
	request, _ = url.Parse("http://monkey.com/test")

	go filter.Test(request)
	go timer(timeSignal, 5*time.Millisecond)

	select {
	case <-filter.Results():
		// Good, response returned first
	case <-timeSignal:
		t.Errorf("Initial response was too slow")
		return
	}

	filter.ProcessResponse(response)
	// Now that the first 500 has returned, the next request should be delayed
	// by over 90ms
	timeSignal = make(chan bool)
	request, _ = url.Parse("http://monkey.com/test1")
	go filter.Test(request)
	go timer(timeSignal, 90*time.Millisecond)
	select {
	case <-filter.Results():
		t.Errorf("First delay was too short")
		return
	case <-timeSignal:
		// Good, at least 90ms should pass
	}
	<-filter.Results()

	filter.ProcessResponse(response)
	filter.ProcessResponse(response)
	// After two more 500 errors, the next request should be
	// delayed by at least 300ms
	timeSignal = make(chan bool)
	request, _ = url.Parse("http://monkey.com/test2")
	go filter.Test(request)
	go timer(timeSignal, 300*time.Millisecond)
	select {
	case <-filter.Results():
		t.Errorf("Second delay was too short")
		return
	case <-timeSignal:
		// Good, at least 90ms should pass
	}
	<-filter.Results()

	// Message to a different domain should pass instantly
	timeSignal = make(chan bool)
	request, _ = url.Parse("http://pie.com/test")
	go filter.Test(request)
	go timer(timeSignal, 5*time.Millisecond)
	select {
	case <-filter.Results():
		// Good, response returned first
	case <-timeSignal:
		t.Errorf("Un-delayed domain was too slow")
		return
	}
}

func TestFiltersBlacklistsDomain(t *testing.T) {
	filter := Create()

	filter.Blacklist("monkey.com")

	go func() {
		defer filter.Close()
		url1, _ := url.Parse("http://monkey.com/test1.html")
		url2, _ := url.Parse("http://pie.com/test1.html")
		filter.Test(url1)
		filter.Test(url2)
	}()

	//First response should pass through as normal
	pieReturned := false
	for response := range filter.Results() {
		if response.Host == "monkey.com" {
			t.Errorf("Blacklisted domain not filtered")
		} else if response.Host == "pie.com" {
			pieReturned = true
		}
	}

	if !pieReturned {
		t.Errorf("non-blacklisted domain incorrectly filtered")
	}
}

func TestFilterBlocksDuplicates(t *testing.T) {
	filter := Create()

	go func() {
		defer filter.Close()
		url1, _ := url.Parse("http://monkey.com/test1.html")
		url2, _ := url.Parse("http://monkey.com/test1.html")
		filter.Test(url1)
		filter.Test(url2)
	}()

	//Only one result should return
	responseCount := 0
	for range filter.Results() {
		responseCount++
	}
	if responseCount > 1 {
		t.Errorf("Duplicate response returned")
	}
}
