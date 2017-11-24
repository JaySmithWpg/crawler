// Package resolver is for resolving the IP addresses of hosts in the Message interface.
package resolver

import (
	"net"
	"sync"
)

// Message interface passed to and from the resolver
type Message interface {
	HostName() string
	Address() net.IP
	SetAddress(net.IP)
	Error() string
	SetError(string)
}

type resolver struct {
	dnsResolver func(domain string) ([]net.IP, error)
	resolved    chan<- Message
	errors      chan<- Message
	cacheLock   sync.RWMutex
	waitGroup   sync.WaitGroup
	cache       map[string]net.IP
}

// Create receives a channel of Messages and resolves their IP address.
// Returns a channel of Messages pointers with their IP address populated.
// Also returns a channel of Messages that encountered errors
func Create(requests <-chan Message) (<-chan Message, <-chan Message) {
	return createInjected(requests, net.LookupIP)
}

func createInjected(requests <-chan Message, dnsResolver func(string) ([]net.IP, error)) (<-chan Message, <-chan Message) {
	resolved := make(chan Message)
	errors := make(chan Message)

	r := &resolver{
		dnsResolver: dnsResolver,
		cacheLock:   sync.RWMutex{},
		cache:       make(map[string]net.IP),
		resolved:    resolved,
		errors:      errors,
	}

	go func() {
		defer close(resolved)
		defer close(errors)
		for request := range requests {
			r.waitGroup.Add(1)
			go r.resolve(request)
		}
		r.waitGroup.Wait()
	}()

	return resolved, errors
}

func (r *resolver) resolve(request Message) {
	defer r.waitGroup.Done()

	r.cacheLock.RLock()
	cachedIP := r.cache[request.HostName()]
	r.cacheLock.RUnlock()
	if cachedIP == nil {
		responseIps, err := r.dnsResolver(request.HostName())

		if err != nil {
			request.SetError(err.Error())
			r.errors <- request
		} else {
			r.cacheLock.Lock()
			r.cache[request.HostName()] = responseIps[0]
			r.cacheLock.Unlock()
			request.SetAddress(responseIps[0])
			r.resolved <- request
		}
	} else {
		request.SetAddress(cachedIP)
		r.resolved <- request
	}
}
