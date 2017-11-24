package parser

import (
	"net/http"
)

type Request interface {
	Body() *[]byte
	Response() *http.Response
}
