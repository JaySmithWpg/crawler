package main

import (
	"errors"
	"net"
	"testing"
)

func TestResolverSuccess(t *testing.T) {
	expectedAddress := net.IPv4(2, 2, 2, 3)

	// stub out the dns resolver
	oldResolver := dnsResolver
	defer func() { dnsResolver = oldResolver }()
	dnsResolver = func(domain string) ([]net.IP, error) {
		if domain == "Monkeyland.com" {
			return []net.IP{expectedAddress}, nil
		}
		return []net.IP{net.IPv6zero}, nil
	}

	resolveRequests := make(chan *downloadRequest)

	go func() {
		defer close(resolveRequests)
		resolveRequests <- &downloadRequest{
			hostName: "Monkeyland.com",
			path:     "/",
			port:     443,
			https:    true,
			newHost:  true,
			address:  net.IPv6zero,
		}
	}()

	resolvedRequests, _ := resolver(resolveRequests)

	resolved := <-resolvedRequests

	if !resolved.address.Equal(expectedAddress) {
		t.Errorf("Expected %s, found %s", expectedAddress.String(), resolved.address.String())
	}

}

func TestResolverFailure(t *testing.T) {
	// stub out the dns resolver
	oldResolver := dnsResolver
	defer func() { dnsResolver = oldResolver }()
	dnsResolver = func(domain string) ([]net.IP, error) {
		return nil, errors.New("Can't Resolve Domain")
	}

	resolveRequests := make(chan *downloadRequest)

	go func() {
		defer close(resolveRequests)
		resolveRequests <- &downloadRequest{
			hostName: "Monkeyland.com",
			path:     "/",
			port:     443,
			https:    true,
			newHost:  true,
			address:  net.IPv6zero,
		}
	}()

	_, failedRequests := resolver(resolveRequests)

	resolverError := <-failedRequests

	if resolverError.Error() != "Monkeyland.com: Can't Resolve Domain" {
		t.Errorf("Unexpected error message returned: %s", resolverError.Error())
	}

	if resolverError.hostName != "Monkeyland.com" {
		t.Errorf("Unexpected domain returned: %s", resolverError.hostName)
	}
}

func TestResolverCache(t *testing.T) {
	expectedAddress := net.IPv4(2, 2, 2, 3)
	resolverCalled := false

	// stub out the dns resolver
	oldResolver := dnsResolver
	defer func() { dnsResolver = oldResolver }()
	dnsResolver = func(domain string) ([]net.IP, error) {
		if resolverCalled {
			t.Errorf("Resolver called too many times")
		}
		resolverCalled = true
		if domain == "Monkeyland.com" {
			return []net.IP{expectedAddress}, nil
		}
		return []net.IP{net.IPv6zero}, nil
	}

	resolveRequests := make(chan *downloadRequest)
	defer close(resolveRequests)

	resolvedRequests, _ := resolver(resolveRequests)
	go func() {
		resolveRequests <- &downloadRequest{
			hostName: "Monkeyland.com",
			path:     "/",
			port:     443,
			https:    true,
			newHost:  true,
			address:  net.IPv6zero,
		}
	}()

	<-resolvedRequests
	go func() {
		resolveRequests <- &downloadRequest{
			hostName: "Monkeyland.com",
			path:     "/",
			port:     443,
			https:    true,
			newHost:  true,
			address:  net.IPv6zero,
		}
	}()
	<-resolvedRequests

}
