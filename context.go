package webgo

import (
	"net/http"
	"encoding/json"
	"fmt"
	"bytes"
	"strings"
	"time"
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


func (c *Context) GetCookie(key string) string {
	val, err := c.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return val.Value
}
// Порядок params - MaxAge, Path, Domain, HttpOnly, Secure
// Внимание! HttpOnly для сессий необходимо передавать true!!! Это органичет доступ к кукам JS в браузере
func (c *Context) SetCookie (name string, val string, params ...interface{}) {
	var cookie bytes.Buffer

	// Очищаем спец символы
	nameCleaner := strings.NewReplacer("\n","-","\r","-")
	name = nameCleaner.Replace(name)

	valueCleaner := strings.NewReplacer("\n"," ","\r"," ",";"," ")
	val = valueCleaner.Replace(val)

	fmt.Fprintf(&cookie, "%s=%s", name, val)

	ln := len(params)

	if ln > 0 {
		var maxAge int64

		switch v := params[0].(type) {
		case int:
			maxAge = int64(v)
		case int32:
			maxAge = int64(v)
		case int64:
			maxAge = v
		}

		if maxAge > 0 {
			fmt.Fprintf(&cookie, "; Expires=%s; Max-Age=%d", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(time.RFC1123), maxAge)
		} else {
			fmt.Fprintf(&cookie, "; Max-Age=0")
		}
	}

	// Устанавливаем Path
	if ln > 1 {
		if v, ok := params[1].(string); ok && v != "" {
			fmt.Fprintf(&cookie, "; Path=%s", valueCleaner.Replace(v))
		}
	} else {
		fmt.Fprintf(&cookie, "; Path=%s", "/")
	}

	// Устанавливаем Domain
	if ln > 2 {
		if v, ok := params[2].(string); ok && v != "" {
			fmt.Fprintf(&cookie, "; Domain=%s", valueCleaner.Replace(v))
		}
	}

	// Устанавливаем HttpOnly
	if ln > 3 {
		if v, ok := params[3].(bool); ok && v {
			fmt.Fprintf(&cookie, "; HttpOnly")
		}
	}

	// Устанавливаем Secure
	if ln > 4 {
		var secure bool
		switch v := params[4].(type) {
		case bool:
			secure = v
		default:
			if params[4] != nil {
				secure = true
			}
		}

		if secure {
			fmt.Fprintf(&cookie, "; Secure")
		}

	}

	c.Response.Header().Add("Set-Cookie", cookie.String())
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
