package resolver

import (
	"crawler/utils"
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

	resolveRequests := make(chan Request)

	go func() {
		defer close(resolveRequests)
		resolveRequests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()

	resolvedRequests, _ := Resolver(resolveRequests)

	resolved := <-resolvedRequests

	if !resolved.Address().Equal(expectedAddress) {
		t.Errorf("Expected %s, found %s", expectedAddress.String(), resolved.Address().String())
	}
}

func TestResolverFailure(t *testing.T) {
	// stub out the dns resolver
	oldResolver := dnsResolver
	defer func() { dnsResolver = oldResolver }()
	dnsResolver = func(domain string) ([]net.IP, error) {
		return nil, errors.New("Can't Resolve Domain")
	}

	resolveRequests := make(chan Request)

	go func() {
		defer close(resolveRequests)
		resolveRequests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()

	_, failedRequests := Resolver(resolveRequests)

	resolverError := <-failedRequests

	if resolverError.Error() != "Monkeyland.com: Can't Resolve Domain" {
		t.Errorf("Unexpected error message returned: %s", resolverError.Error())
	}

	if resolverError.HostName != "Monkeyland.com" {
		t.Errorf("Unexpected domain returned: %s", resolverError.HostName)
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

	resolveRequests := make(chan Request)
	defer close(resolveRequests)

	resolvedRequests, _ := Resolver(resolveRequests)
	go func() {
		resolveRequests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()

	<-resolvedRequests
	go func() {
		resolveRequests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()
	<-resolvedRequests
}
