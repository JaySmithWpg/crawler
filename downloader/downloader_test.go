package downloader

import (
	"bytes"
	"fmt"
	"github.com/JaySmithWpg/crawler/utils"
	"net"
	"net/http"
	"testing"
	"time"
)

//TODO: Stub out the crawler/utils dependency

func makeDownloadMessage(url string, addr net.IP) Message {
	request := utils.CreateCrawlerRecord(url)
	request.SetAddress(addr)
	return request
}

func serveTestPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Host == "monkeys.com:9090" || r.Host == "monkeys.com:9091" {
		fmt.Fprintf(w, "Hello World!") // send data to client side
	}
}

func TestDownloader(t *testing.T) {
	server := &http.Server{Addr: "127.0.0.1:9090"}
	http.HandleFunc("/monkey.html", serveTestPage)
	go func() {
		server.ListenAndServe()
	}()
	//Give the server time to start listening
	time.Sleep(100 * time.Millisecond)

	downloadMessages := make(chan Message)
	go func() {
		defer close(downloadMessages)
		downloadMessages <- makeDownloadMessage("http://monkeys.com:9090/monkey.html", net.IPv4(127, 0, 0, 1))
	}()

	completed, _ := Create(downloadMessages)

	response := <-completed
	if bytes.Compare(*response.Body(), []byte("Hello World!")) != 0 {
		t.Errorf("Unexpected Body: ", response.Body())
	}
}

func TestHttpTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running timeout test in short mode")
	}
	t.Parallel()

	msg := makeDownloadMessage("http://monkeys.com/", net.IPv4(192, 168, 10, 1))

	downloadMessages := make(chan Message)
	go func() {
		defer close(downloadMessages)
		downloadMessages <- msg
	}()

	_, failed := Create(downloadMessages)

	fail := <-failed
	if fail.Error() != "dial tcp 192.168.10.1:80: getsockopt: connection timed out" {
		t.Errorf("Wrong error message: %s", fail.Error())
	}
}

func TestHttpsTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running timeout test in short mode")
	}
	t.Parallel()

	msg := makeDownloadMessage("https://monkeys.com/", net.IPv4(192, 168, 10, 1))

	downloadMessages := make(chan Message)
	go func() {
		defer close(downloadMessages)
		downloadMessages <- msg
	}()

	_, failed := Create(downloadMessages)

	fail := <-failed
	if fail.Error() != "dial tcp 192.168.10.1:443: getsockopt: connection timed out" {
		t.Errorf("Wrong error message: %s", fail.Error())
	}
}
