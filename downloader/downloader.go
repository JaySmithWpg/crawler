package downloader

import (
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Max retries before giving up on a page
const maxRetries = 3

type Downloader interface {
	Request(*url.URL)
	Completed() <-chan *http.Response
	Failed() <-chan error
	Close()
}

type downloader struct {
	requests        chan *url.URL
	downloaded      chan *http.Response
	failedDownloads chan error
}

// Create accepts a channel of messages containing download instructions
// Create returns a channel of successful downloads and a channel of failed downloads
func Create() Downloader {
	d := &downloader{
		requests:        make(chan *url.URL),
		downloaded:      make(chan *http.Response),
		failedDownloads: make(chan error),
	}
	go d.run()
	return d
}

func (d *downloader) Request(u *url.URL) {
	d.requests <- u
}

func (d *downloader) Completed() <-chan *http.Response {
	return d.downloaded
}

func (d *downloader) Failed() <-chan error {
	return d.failedDownloads
}

func (d *downloader) Close() {
	close(d.requests)
}

func (d *downloader) run() {
	defer close(d.downloaded)
	defer close(d.failedDownloads)

	// Used for random jitter when resending requests
	rand.Seed(time.Now().Unix())

	var wg sync.WaitGroup
	for request := range d.requests {
		wg.Add(1)
		go d.process(request, &wg)
	}
	wg.Wait()
}

func (d *downloader) process(u *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	httpResponse, err := send(u)
	if err == nil {
		d.downloaded <- httpResponse
	} else {
		d.failedDownloads <- err
	}
}

func send(u *url.URL) (*http.Response, error) {
	//HACK: I used to do local domain caching, but have since decided that it's easier
	//     to just install something like Unbound to do it system-wide for me
	var err error
	var httpResponse *http.Response

	for i := 0; i < maxRetries; i++ {
		httpResponse, err = http.Get(u.String())
		if err == nil {
			return httpResponse, nil
		}
		//exponential backoff with some random jitter
		delay := int(math.Pow(200, float64(i)))
		time.Sleep(time.Duration(delay+rand.Intn(60)) * time.Millisecond)
	}

	return nil, err
}
