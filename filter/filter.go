// The filter will not let duplicate urls or urls that violate robots.txt pass.
package filter

import (
	"net/http"
	"net/url"
	"sync"
)

// Urls will be partitioned by host, removing the need for locks around the host state.
// One goroutine per host, not per url.
// HACK: maps may not be the most efficient data structure for urlsSeen,
//       but we'll try it first before decided if something like a red-black tree
//       is required.
type host struct {
	blacklist chan bool
	urlsSeen  map[string]bool
	requests  chan *url.URL
	response  chan<- *url.URL
}

func (h *host) isValid(u *url.URL) bool {
	isRepeat := h.urlsSeen[u.String()]
	if isRepeat {
		return false
	}

	h.urlsSeen[u.String()] = true
	return true
}

func (h *host) run(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(h.blacklist)

	blacklisted := false
	for {
		select {
		case u, ok := <-h.requests:
			if !ok {
				return
			}
			if !blacklisted && h.isValid(u) {
				h.response <- u
			}
		case blacklisted = <-h.blacklist:
		}
	}
}

type Filter interface {
	Test(*url.URL)
	Results() <-chan *url.URL
	BlacklistHost(string)
	Close()
}

type filter struct {
	requests      chan *url.URL
	results       chan *url.URL
	hosts         map[string]*host
	blacklist     chan string
	hostWaitGroup sync.WaitGroup
}

func Create() Filter {
	f := &filter{
		requests:      make(chan *url.URL),
		results:       make(chan *url.URL),
		hosts:         make(map[string]*host),
		blacklist:     make(chan string),
		hostWaitGroup: sync.WaitGroup{},
	}
	go f.run()
	return f
}

func (f *filter) Test(u *url.URL) {
	f.requests <- u
}

func (f *filter) Results() <-chan *url.URL {
	return f.results
}

func (f *filter) BlacklistHost(host string) {
	f.blacklist <- host
}

func (f *filter) Close() {
	close(f.requests)
}

func (f *filter) getHost(hst string) *host {
	h, hostExists := f.hosts[hst]
	if !hostExists {
		// Using small buffered channels to communicate with the hosts
		// so delays in processing one host's request will not block the
		// main filter when a subsequent request is sent to the same host
		h = &host{
			blacklist: make(chan bool),
			requests:  make(chan *url.URL, 5),
			response:  f.results,
			urlsSeen:  make(map[string]bool),
		}
		f.hosts[hst] = h
		f.hostWaitGroup.Add(1)
		go h.run(&f.hostWaitGroup)
	}
	return h
}

func (f *filter) closeHosts() {
	// I don't use a shared `done` channel to close these all at once,
	// because the hosts have buffered channels that they must finish
	// processing without interruption before shutting down. Closing
	// the channels will allow them to read the channel buffer until
	// it's empty before shutting down.
	for _, host := range f.hosts {
		close(host.requests)
	}
	f.hostWaitGroup.Wait()
}

func (f *filter) run() {
	defer close(f.blacklist)
	defer close(f.results)

	for {
		select {
		case u, ok := <-f.requests:
			if !ok {
				f.closeHosts()
				return
			}
			host := f.getHost(u.Host)
			host.requests <- u
		case hostKey := <-f.blacklist:
			host := f.getHost(hostKey)
			host.blacklist <- true
		}
	}
}
