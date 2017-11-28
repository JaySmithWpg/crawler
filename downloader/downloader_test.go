package downloader

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

//TODO: this is entirely unneeded
func makeDownloadMessage(urlString string, addr string) *url.URL {
	u, _ := url.Parse(urlString)
	return u
}

func serveTestPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprintf(w, "Hello World!") // send data to client side
}

func TestDownloader(t *testing.T) {
	server := &http.Server{Addr: "127.0.0.1:9090"}
	http.HandleFunc("/monkey.html", serveTestPage)
	go func() {
		server.ListenAndServe()
	}()
	//Give the server time to start listening
	time.Sleep(100 * time.Millisecond)

	d := Create()
	go func() {
		defer d.Close()
		d.Request(makeDownloadMessage("http://localhost:9090/monkey.html", "127.0.0.1:9090"))
	}()

	r := <-d.Completed()
	if r.StatusCode != 200 {
		t.Errorf("Bad Response Code Returned: %d", r.StatusCode)
	}
}

func TestHttpTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running timeout test in short mode")
	}
	t.Parallel()

	msg := makeDownloadMessage("http://localhost:10000/", "192.168.10.1:80")

	d := Create()
	go func() {
		defer d.Close()
		d.Request(msg)
	}()

	fail := <-d.Failed()
	if fail.Error() != "Get http://localhost:10000/: dial tcp [::1]:10000: i/o timeout" {
		t.Errorf("Wrong error message: %s", fail.Error())
	}
}

func TestHttpsTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running timeout test in short mode")
	}
	t.Parallel()

	msg := makeDownloadMessage("https://localhost:10000/", "192.168.10.1:443")

	d := Create()
	go func() {
		defer d.Close()
		d.Request(msg)
	}()

	fail := <-d.Failed()
	if fail.Error() != "Get https://localhost:10000/: dial tcp [::1]:10000: i/o timeout" {
		t.Errorf("Wrong error message: %s", fail.Error())
	}
}
