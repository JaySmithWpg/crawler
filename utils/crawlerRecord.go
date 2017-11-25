package utils

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CrawlerRecord interface {
	Url() *url.URL
	HostName() string
	Path() string
	Port() uint16
	IsHttps() bool
	HasResponse() bool
	Address() string
	SetAddress(net.IP)
	SetResponse(*http.Response)
	Body() *[]byte
	Headers() map[string][]string
	Error() string
	SetError(string)
}
type crawlerRecord struct {
	url         *url.URL
	path        string
	port        uint16
	isHttps     bool
	hasResponse bool
	address     net.IP
	body        []byte
	headers     map[string][]string
	error       string
}

func (c *crawlerRecord) Url() *url.URL {
	return c.url
}

func (c *crawlerRecord) HostName() string {
	return c.url.Hostname()
}

func (c *crawlerRecord) Path() string {
	if c.url.Path != "" {
		return c.url.Path
	} else {
		return "/"
	}
}

func (c *crawlerRecord) Port() uint16 {
	return c.port
}

func (c *crawlerRecord) IsHttps() bool {
	return c.isHttps
}

func (c *crawlerRecord) Address() string {
	return fmt.Sprintf("%s:%d", c.address.String(), c.Port())
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
	if r.ContentLength > 0 {
		c.body = make([]byte, r.ContentLength)
		r.Body.Read(c.body)
		r.Body.Close()
	}
	c.headers = r.Header
	c.hasResponse = true
}

func (c *crawlerRecord) Error() string {
	return c.error
}

func (c *crawlerRecord) SetError(e string) {
	c.error = e
}

func CreateRecordFromUrl(u *url.URL) CrawlerRecord {
	isHttps := strings.HasPrefix(strings.ToLower(u.Scheme), "https")

	port, err := strconv.ParseUint(u.Port(), 10, 16)

	if err != nil {
		if isHttps {
			port = 443
		} else {
			port = 80
		}
	}

	return &crawlerRecord{
		url:     u,
		isHttps: isHttps,
		port:    uint16(port),
	}
}

//Take an url and turns it into a record for crawling
func CreateCrawlerRecord(s string) CrawlerRecord {
	u, _ := url.Parse(s)
	return CreateRecordFromUrl(u)
}
