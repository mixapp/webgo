package webgo

import (
	"net/http"
)

type Context struct {
	Response http.ResponseWriter
	Request *http.Request
	Output interface{}
	Query map[interface{}]interface{}
	Body map[interface{}]interface{}
	statusCode int
	body []byte
}