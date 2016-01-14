package webgo

import (
	"net/http"
)

type Context struct {
	Response http.ResponseWriter
	Request *http.Request
	Error error
	Output interface{}
	statusCode int
	body []byte
}