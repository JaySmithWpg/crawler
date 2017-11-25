package resolver

import (
	"errors"
	"net"
	"testing"
)

//TODO: Find a framework to automatically mock this
type message struct {
	address  net.IP
	hostName string
	error    string
}

func (m *message) HostName() string {
	return m.hostName
}

func (m *message) Address() string {
	return m.address.String()
}

func (m *message) SetAddress(a net.IP) {
	m.address = a
}

func (m *message) Error() string {
	return m.error
}

func (m *message) SetError(s string) {
	m.error = s
}

func TestResolverSuccess(t *testing.T) {
	expectedAddress := net.IPv4(2, 2, 2, 3)
	msg := &message{hostName: "Monkeyland.com"}

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
		requests <- msg
	}()

	resolved := <-resolvedMessages

	if resolved.Address() != expectedAddress.String() {
		t.Errorf("Expected %s, found %s", expectedAddress.String(), resolved.Address())
	}
}

func TestResolverFailure(t *testing.T) {
	// stub out the dns resolver
	dnsResolver := func(domain string) ([]net.IP, error) {
		return nil, errors.New("Can't Resolve Domain")
	}
	msg := &message{hostName: "Monkeyland.com"}
	requests := make(chan Message)

	go func() {
		defer close(requests)
		requests <- msg
	}()

	_, failedMessages := createInjected(requests, dnsResolver)
	fail := <-failedMessages

	if fail.Error() != "Can't Resolve Domain" {
		t.Errorf("Unexpected error message returned: %s", fail.Error())
	}

	if fail.HostName() != "Monkeyland.com" {
		t.Errorf("Unexpected domain returned: %s", fail.HostName())
	}
}

func TestResolverCache(t *testing.T) {
	expectedAddress := net.IPv4(2, 2, 2, 3)
	resolverCalled := false
	msg1 := &message{hostName: "Monkeyland.com"}
	msg2 := &message{hostName: "Monkeyland.com"}

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
		requests <- msg1
	}()

	<-resolvedMessages
	go func() {
		requests <- msg2
	}()
	<-resolvedMessages
}
