package utils

import (
	"bytes"
	"net"
	"net/http"
	"reflect"
	"testing"
)

func TestAddressPort(t *testing.T) {
	var r CrawlerRecord
	var addr net.IP

	r = CreateCrawlerRecord("https://monkeys.com:65535/pie.html")
	addr = net.IPv4(2, 2, 2, 2)
	r.SetAddress(addr)
	if r.Address() != "2.2.2.2:65535" {
		t.Errorf("Incorrect Address: %s", r.Address())
	}

	r = CreateCrawlerRecord("http://monkeys.com/pie.html")
	addr = net.IPv4(2, 2, 2, 2)
	r.SetAddress(addr)
	if r.Address() != "2.2.2.2:80" {
		t.Errorf("Incorrect Address: %s", r.Address())
	}

	r = CreateCrawlerRecord("https://monkeys.com/pie.html")
	addr = net.IPv4(2, 2, 2, 2)
	r.SetAddress(addr)
	if r.Address() != "2.2.2.2:443" {
		t.Errorf("Incorrect Address: %s", r.Address())
	}
}

func TestCreateCrawlerRecord(t *testing.T) {
	r := CreateCrawlerRecord("https://monkeys.com:3444/banana/pie.html")

	hostTest := r.HostName()
	if hostTest != "monkeys.com" {
		t.Errorf("Incorrect host name: %s", hostTest)
	}

	pathTest := r.Path()
	if pathTest != "/banana/pie.html" {
		t.Errorf("Incorrect path: %s", pathTest)
	}

	if r.Port() != 3444 {
		t.Errorf("Incorrect port.")
	}

	if r.HasResponse() {
		t.Errorf("HasResponse lies!")
	}

	if !r.IsHttps() {
		t.Errorf("Https flag is incorrect")
	}

	if CreateCrawlerRecord("http://pie.com/").Port() != 80 || CreateCrawlerRecord("https://pie.com/").Port() != 443 {
		t.Errorf("Wrong default port")
	}

	rootPath := CreateCrawlerRecord("http://pie.com").Path()
	if rootPath != "/" {
		t.Errorf("Root path not correctly set: %s", rootPath)
	}
}

type bodyStub struct {
	data []byte
}

func (b bodyStub) Read(p []byte) (int, error) {
	copy(p, b.data)
	return len(p), nil
}

func (b bodyStub) Close() error {
	return nil
}

func TestUrl(t *testing.T) {
	record := CreateCrawlerRecord("https://banana.com/pie/apple/orange.html")
	if record.Url().String() != "https://banana.com/pie/apple/orange.html" {
		t.Errorf("URL not set correctly")
	}
}
func TestSetResponse(t *testing.T) {
	record := CreateCrawlerRecord("https://monkeys.com")

	header := http.Header{}
	header.Add("foo", "bar")
	header.Add("foo", "2")
	body := bodyStub{data: []byte("This is some body text")}

	response := &http.Response{
		ContentLength: int64(len(body.data)),
		Header:        header,
		Body:          body,
	}

	record.SetResponse(response)
	if !bytes.Equal(*record.Body(), []byte("This is some body text")) {
		t.Errorf("Body value incorrect: %s", string(*record.Body()))
	}

	if !reflect.DeepEqual(header["foo"], record.Headers()["foo"]) {
		t.Errorf("HTTP response headers not correctly set: %s\n Should be: %s", record.Headers(), header)
	}

	if !record.HasResponse() {
		t.Errorf("HasResponse flag not set")
	}
}
