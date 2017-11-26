// filter serves as the brains of the crawler.
// The filter will not let duplicate urls or urls that violate robots.txt pass.
// The filter will delay requests that threaten to overwhelm hosts
package filter

import (
	"net/url"
	"sync"
)

// we might want something like a red-black tree for performance in the future
// but for now, lets use the built in maps
type filter struct {
	urls    map[string]bool
	urlLock sync.RWMutex
}

func (f *filter) isValid(u *url.URL) bool {

	f.urlLock.RLock()
	urlSeenBefore := f.urls[u.String()]
	f.urlLock.RUnlock()

	if urlSeenBefore {
		return false
	}

	f.urlLock.Lock()
	//Always check again after locking
	urlSeenBefore = f.urls[u.String()]
	if !urlSeenBefore {
		f.urls[u.String()] = true
	}
	f.urlLock.Unlock()
	return !urlSeenBefore
}

// Create takes a channel of urls, and returns a
// channel of only urls that are safe to download
func Create(requests <-chan *url.URL) <-chan *url.URL {
	responses := make(chan *url.URL)

	f := &filter{
		urls:    make(map[string]bool),
		urlLock: sync.RWMutex{},
	}

	var wg sync.WaitGroup
	go func() {
		defer close(responses)
		for r := range requests {
			wg.Add(1)
			go func(r *url.URL) {
				defer wg.Done()
				if f.isValid(r) {
					responses <- r
				}
			}(r)

		}
		wg.Wait()
	}()
	return responses
}
