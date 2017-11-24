package downloader

import (
	"bytes"
	"crawler/utils"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

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
		downloadMessages <- makeDownloadMessage("http://monkysrstr.co:8434.com/", net.IPv4(192, 168, 10, 1))
	}()

	completed, failed := Create(downloadMessages)

	response := <-completed
	if bytes.Compare(*response.Body(), []byte("Hello World!")) != 0 {
		t.Errorf("Unexpected Body: ", response.Body())
	}

	failure := <-failed
	if failure.HostName() != "monkysrstr.co" {
		t.Errorf("Wrong domain failed: %s", failure.HostName())
	}
	if failure.Error() != "dial tcp 192.168.10.1:0: i/o timeout" {
		t.Errorf("Wrong error message: %s", failure.Error())
	}

}
