// Package for resolving the IP addresses of hosts in the Request interface.
package resolver

import (
	"fmt"
	"net"
	"sync"
)

//TODO: NO. Update to no longer use a worker thread model.
const resolverThreads = 5

// The following values can be stubbed out during testing
var dnsResolver = net.LookupIP

//TODO: Update error handling to be consistent with downloader 
type Error struct {
	HostName    string
	Description string
}

type Request interface {
	HostName() string
	Address() net.IP
	SetAddress(net.IP)
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.HostName, e.Description)
}

// Receives a channel of DownloadRequest pointers and resolves their IP address.
// Returns a channel of DownloadRequest pointers with their IP address populated.
// Also returns a channel of Errors
func Resolver(requests <-chan Request) (<-chan Request, <-chan *Error) {
	resolved := make(chan Request)
	errors := make(chan *Error)

	//TODO: proper cache that expires entries rather than eat all memory
	cache := make(map[string]net.IP)
	lock := &sync.RWMutex{}

	go func() {
		defer close(resolved)
		defer close(errors)
		var wg sync.WaitGroup
		for request := range requests {
			wg.Add(1)
			go resolveWorker(request, resolved, errors, cache, lock, &wg)
		}
		wg.Wait()
	}()
	return resolved, errors
}

func resolveWorker(request Request, results chan<- Request, errors chan<- *Error, cache map[string]net.IP, lock *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.RLock()
	cachedIP := cache[request.HostName()]
	lock.RUnlock()
	if cachedIP == nil {
		responseIps, err := dnsResolver(request.HostName())

		if err != nil {
			errors <- &Error{HostName: request.HostName(), Description: err.Error()}
		} else {
			lock.Lock()
			cache[request.HostName()] = responseIps[0]
			lock.Unlock()
			request.SetAddress(responseIps[0])
			results <- request
		}
	} else {
		request.SetAddress(cachedIP)
		results <- request
	}
}
