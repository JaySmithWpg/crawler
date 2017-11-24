package utils

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

type CrawlerRecord interface {
	HostName() string
	Path() string
	Port() int
	IsHttps() bool
	HasResponse() bool
	Address() net.IP
	SetAddress(net.IP)
	SetResponse(*http.Response)
	Body() *[]byte
	Headers() map[string][]string
	Error() string
	SetError(string)
}
type crawlerRecord struct {
	hostName    string
	path        string
	port        int
	isHttps     bool
	hasResponse bool
	address     net.IP
	body        []byte
	headers     map[string][]string
	error       string
}

func (c *crawlerRecord) HostName() string {
	return c.hostName
}

func (c *crawlerRecord) Path() string {
	return c.path
}

func (c *crawlerRecord) Port() int {
	return c.port
}

func (c *crawlerRecord) IsHttps() bool {
	return c.isHttps
}

func (c *crawlerRecord) Address() net.IP {
	return c.address
}

func (c *crawlerRecord) SetAddress(a net.IP) {
	c.address = a
}

func (c *crawlerRecord) Body() *[]byte {
	return &c.body
}

func (c *crawlerRecord) Headers() map[string][]string {
	return c.headers
}

func (c *crawlerRecord) HasResponse() bool {
	return c.hasResponse
}

func (c *crawlerRecord) SetResponse(r *http.Response) {
	c.body = make([]byte, r.ContentLength)
	r.Body.Read(c.body)
	r.Body.Close()
	c.headers = r.Header
	c.hasResponse = true
}

func (c *crawlerRecord) Error() string {
	return c.error
}

func (c *crawlerRecord) SetError(e string) {
	c.error = e
}

//Take an url and turns it into a record for crawling
func CreateCrawlerRecord(url string) CrawlerRecord {
	isHttps := strings.HasPrefix(strings.ToLower(url), "https")

	var domainStart int
	if isHttps {
		domainStart = 8
	} else {
		domainStart = 7
	}

	split := strings.SplitN(url[domainStart:len(url)], "/", 2)
	hostDetails := strings.Split(split[0], ":")

	var port int
	if len(hostDetails) < 2 {
		if isHttps {
			port = 443
		} else {
			port = 80
		}
	} else {
		port, _ = strconv.Atoi(hostDetails[1])
	}

	var path string
	if len(split) < 2 {
		path = "/"
	} else {
		path = "/" + split[1]
	}

	return &crawlerRecord{
		isHttps:  isHttps,
		hostName: hostDetails[0],
		port:     port,
		path:     path,
	}
}
