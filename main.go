package main

import (
	"fmt"
	"net"
	"sync"
)

const resolverThreads = 5

// The following values can be stubbed out during testing
var dnsResolver = net.LookupIP

type downloadRequest struct {
	hostName string
	path     string
	port     int
	https    bool
	newHost  bool
	address  net.IP
}

type resolverError struct {
	hostName    string
	description string
}

func (e *resolverError) Error() string {
	return fmt.Sprintf("%s: %s", e.hostName, e.description)
}

func resolver(requests <-chan *downloadRequest) (<-chan *downloadRequest, <-chan *resolverError) {
	resolved := make(chan *downloadRequest)
	errors := make(chan *resolverError)

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

func resolveWorker(request *downloadRequest, results chan<- *downloadRequest, errors chan<- *resolverError, cache map[string]net.IP, lock *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()

	lock.RLock()
	cachedIP := cache[request.hostName]
	lock.RUnlock()
	if cachedIP == nil {
		responseIps, err := dnsResolver(request.hostName)

		if err != nil {
			errors <- &resolverError{hostName: request.hostName, description: err.Error()}
		} else {
			lock.Lock()
			cache[request.hostName] = responseIps[0]
			lock.Unlock()
			request.address = responseIps[0]
			results <- request
		}
	} else {
		request.address = cachedIP
		results <- request
	}
}

func main() {
	var wg sync.WaitGroup
	requests := make(chan *downloadRequest)
	results, errors := resolver(requests)

	wg.Add(2)
	go func() {
		defer wg.Done()
		for result := range results {
			fmt.Printf("%s: %s\n", result.hostName, result.address.String())
		}
	}()

	go func() {
		defer wg.Done()
		for err := range errors {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}()

	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.gooffe.com"}
	requests <- &downloadRequest{hostName: "www.gotogle.com"}
	requests <- &downloadRequest{hostName: "www.godasdogle.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.godarogle.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.gdardsoogle.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.amazon.ca"}
	requests <- &downloadRequest{hostName: "www.amazon.ca"}
	requests <- &downloadRequest{hostName: "www.amazon.ca"}
	requests <- &downloadRequest{hostName: "www.amazon.ca"}
	requests <- &downloadRequest{hostName: "www.amazon.com"}
	requests <- &downloadRequest{hostName: "www.amazon.ca"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	requests <- &downloadRequest{hostName: "www.google.com"}
	close(requests)
	wg.Wait()
}
