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
	dnsResolver := func(domain string) ([]net.IP, error) {
		if domain == "Monkeyland.com" {
			return []net.IP{expectedAddress}, nil
		}
		return []net.IP{net.IPv6zero}, nil
	}
	requests := make(chan Message)
	resolvedMessages, _ := createInjected(requests, dnsResolver)

	go func() {
		defer close(requests)
		requests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()

	resolved := <-resolvedMessages

	if !resolved.Address().Equal(expectedAddress) {
		t.Errorf("Expected %s, found %s", expectedAddress.String(), resolved.Address().String())
	}
}

func TestResolverFailure(t *testing.T) {
	// stub out the dns resolver
	dnsResolver := func(domain string) ([]net.IP, error) {
		return nil, errors.New("Can't Resolve Domain")
	}

	requests := make(chan Message)

	go func() {
		defer close(requests)
		requests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()

	_, failedMessages := createInjected(requests, dnsResolver)
	resolverError := <-failedMessages

	if resolverError.Error() != "Can't Resolve Domain" {
		t.Errorf("Unexpected error message returned: %s", resolverError.Error())
	}

	if resolverError.HostName() != "Monkeyland.com" {
		t.Errorf("Unexpected domain returned: %s", resolverError.HostName())
	}
}

func TestResolverCache(t *testing.T) {
	expectedAddress := net.IPv4(2, 2, 2, 3)
	resolverCalled := false

	// stub out the dns resolver
	dnsResolver := func(domain string) ([]net.IP, error) {
		if resolverCalled {
			t.Errorf("Resolver called too many times")
		}
		resolverCalled = true
		if domain == "Monkeyland.com" {
			return []net.IP{expectedAddress}, nil
		}
		return []net.IP{net.IPv6zero}, nil
	}

	requests := make(chan Message)
	defer close(requests)

	resolvedMessages, _ := createInjected(requests, dnsResolver)
	go func() {
		requests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()

	<-resolvedMessages
	go func() {
		requests <- utils.CreateCrawlerRecord("http://Monkeyland.com")
	}()
	<-resolvedMessages
}
