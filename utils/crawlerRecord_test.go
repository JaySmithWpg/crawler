package utils

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"
)

func TestCreateCrawlerRecord(t *testing.T) {
	r := CreateCrawlerRecord("https://monkeys.com:3444/banana/pie.html")

	if r.HostName() != "monkeys.com" {
		t.Errorf("Incorrect host name.")
	}

	if r.Path() != "/banana/pie.html" {
		t.Errorf("Incorrect path.")
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

	if CreateCrawlerRecord("http://pie.com").Path() != "/" {
		t.Errorf("Root path not correctly set")
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
}
