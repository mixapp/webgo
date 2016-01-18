package webgo

import (
	"net/http"
	"encoding/json"
)

type Context struct {
	Response http.ResponseWriter
	Request *http.Request
	Output interface{}
	Query map[string]interface{}
	_Body []byte
	Body map[string]interface{}
	statusCode int
	body []byte
	Method string
	ContentType string
	error error
}

func (c *Context) ValidateSchema (schema interface{}) (err error) {

	if c.ContentType == "application/x-www-form-urlencoded" {
		// TODO: Реализовать
		return
	}

	if c.ContentType == "application/json" {
		err = json.Unmarshal(c._Body, schema)
		if err != nil {
			return
		}
		return
	}

	return
}

func (c *Context) isString(val interface{}) bool {
	return false
}
func (c *Context) isInteger(val interface{}) bool {
	return false
}
func (c *Context) isMap(val interface{}) bool {
	return false
}
func (c *Context) isSlice(val interface{}) bool {
	return false
}
func (c *Context) isBool(val interface{}) bool {
	return false
}
