package downloader

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

//TODO: logic for downloading robots.txt

// Message interface is used for communication with the downloader
type Message interface {
	HostName() string
	Path() string
	Port() int
	IsHttps() bool
	Address() net.IP
	SetResponse(*http.Response)
	Body() *[]byte
	Headers() map[string][]string
	Error() string
	SetError(e string)
}

// Create accepts a channel of messages containing download instructions
// Create returns a channel of successful downloads and a channel of failed downloads
func Create(requests <-chan Message) (<-chan Message, <-chan Message) {
	downloaded := make(chan Message)
	failedDownloads := make(chan Message)

	go func() {
		defer close(downloaded)
		defer close(failedDownloads)

		var wg sync.WaitGroup
		for request := range requests {
			wg.Add(1)
			go process(request, downloaded, failedDownloads, &wg)
		}
		wg.Wait()
	}()
	return downloaded, failedDownloads
}

func process(r Message, downloaded chan<- Message, toErr chan<- Message, wg *sync.WaitGroup) {
	defer wg.Done()
	httpResponse, err := send(r)
	if err == nil {
		r.SetResponse(httpResponse)
		downloaded <- r
	} else {
		r.SetError(err.Error())
		toErr <- r
	}
	//todo: error reporting
}

func send(r Message) (*http.Response, error) {
	var conn net.Conn
	var err error

	//TODO: back off from domains that are timing out
	//TODO: HTTPS Timeout
	if r.IsHttps() {
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", r.Address().String(), r.Port()), config)

	} else {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", r.Address().String(), r.Port()), 5*time.Second)
	}

	//Direct TCP connection for the request. The HTTP libraries would try to resolve
	//the domain for us - we don't want that.
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(conn, "GET %s HTTP/1.0\r\nHost: %s:%d\r\n\r\n", r.Path(), r.HostName(), r.Port())

	//Using the built in HTTP libraries to parse the response
	httpResponse, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return nil, err
	}
	return httpResponse, nil
}
