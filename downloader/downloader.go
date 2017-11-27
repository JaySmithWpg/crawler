package downloader

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
)

//TODO: logic for downloading robots.txt

// Message interface is used for communication with the downloader
type Message interface {
	Url() *url.URL
	Address() string
	SetResponse(*http.Response)
	Response() *http.Response
	SetError(string)
	Error() string
}

type message struct {
	url      *url.URL
	response *http.Response
	address  string
	error    string
}

func (m *message) Url() *url.URL {
	return m.url
}

func (m *message) Address() string {
	return m.address
}

func (m *message) SetResponse(r *http.Response) {
	m.response = r
}

func (m *message) Response() *http.Response {
	return m.response
}

func (m *message) SetError(e string) {
	m.error = e
}

func (m *message) Error() string {
	return m.error
}

func CreateMessage(u *url.URL, a string) Message {
	return &message{
		url:     u,
		address: a,
	}
}

type Downloader interface {
	Request(Message)
	Completed() <-chan Message
	Failed() <-chan Message
	Close()
}

type downloader struct {
	requests        chan Message
	downloaded      chan Message
	failedDownloads chan Message
}

// Create accepts a channel of messages containing download instructions
// Create returns a channel of successful downloads and a channel of failed downloads
func Create() Downloader {
	d := &downloader{
		requests:        make(chan Message),
		downloaded:      make(chan Message),
		failedDownloads: make(chan Message),
	}
	go d.run()
	return d
}

func (d *downloader) Request(m Message) {
	d.requests <- m
}

func (d *downloader) Completed() <-chan Message {
	return d.downloaded
}

func (d *downloader) Failed() <-chan Message {
	return d.failedDownloads
}

func (d *downloader) Close() {
	close(d.requests)
}

func (d *downloader) run() {
	defer close(d.downloaded)
	defer close(d.failedDownloads)

	var wg sync.WaitGroup
	for request := range d.requests {
		wg.Add(1)
		go d.process(request, &wg)
	}
	wg.Wait()
}

func (d *downloader) process(r Message, wg *sync.WaitGroup) {
	defer wg.Done()
	httpResponse, err := send(r)
	if err == nil {
		r.SetResponse(httpResponse)
		d.downloaded <- r
	} else {
		r.SetError(err.Error())
		d.failedDownloads <- r
	}
}

func send(r Message) (*http.Response, error) {
	var conn net.Conn
	var err error
	//TODO: back off from domains that are timing out
	if r.Url().Scheme == "https" {
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", r.Address(), config)

	} else {
		conn, err = net.Dial("tcp", r.Address())
	}
	//Direct TCP connection for the request. The HTTP libraries would try to resolve
	//the domain for us - we don't want that.
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(conn, "GET %s HTTP/1.0\r\nHost: %s\n\n", r.Url().Path, r.Url().Host)

	//Using the built in HTTP libraries to parse the response
	httpResponse, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return nil, err
	}
	return httpResponse, nil
}
