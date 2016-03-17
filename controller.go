package webgo

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type (
	Controller struct {
		Ctx *Context
		Action string
	}
	ControllerInterface interface {
		Init(ctx *Context, route *Match)
		Prepare() bool
		Finish()
		Error(code int, tpl string)

		GetHeader(key string) string
		SetHeader(key string, val string)
		SetStatusCode(code int)

		Redirect(location string, code int)

		Render(tpl_name string, data interface{})
		Json(data interface{}, unicode bool)
		Plain(data string)
	}
)

func (c *Controller) Init(ctx *Context, route *Match) {
	c.Ctx = ctx
	c.Action = route.Options.Action
}
func (c Controller) Prepare() bool {
	return true
}
func (c Controller) Finish() {}
func (c Controller) Error(code int, data string) {
	http.Error(c.Ctx.Response, data, code)
}

func (c Controller) GetHeader(key string) string {
	return c.Ctx.Request.Header.Get(key)
}
func (c Controller) SetHeader(key string, val string) {
	c.Ctx.Response.Header().Set(key, val)
}
func (c *Controller) SetStatusCode(code int) {
	c.Ctx.statusCode = code
}

func (c Controller) Redirect(location string, code int) {
	http.Redirect(c.Ctx.Response, c.Ctx.Request, location, code)
}

func (c Controller) Render(tpl_name string, data interface{}) {
	var err error

	if c.Ctx.statusCode != 0 {
		c.Ctx.Response.WriteHeader(c.Ctx.statusCode)
	}

	bytes := bytes.NewBufferString("")
	err = app.templates.ExecuteTemplate(bytes, tpl_name+".html", data)
	if err != nil {
		c.Ctx.error = err
	}
	c.Ctx.body, err = ioutil.ReadAll(bytes)
	c.Ctx.Response.Write(c.Ctx.body)
}

func (c Controller) Json(data interface{}, unicode bool) {
	if c.Ctx.statusCode != 0 {
		c.Ctx.Response.WriteHeader(c.Ctx.statusCode)
	}

	c.Ctx.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	var content []byte
	content, err := json.Marshal(data)
	if err != nil {
		c.Ctx.error = err
	}

	if !unicode {
		c.Ctx.Response.Write(content)
		return
	}

	rs := []rune(string(content))
	jsons := ""
	for _, r := range rs {
		rint := int(r)
		if rint < 128 {
			jsons += string(r)
		} else {
			jsons += "\\u" + strconv.FormatInt(int64(rint), 16)
		}
	}

	c.Ctx.Response.Write([]byte(jsons))
}
func (c Controller) Plain(data string) {
	if c.Ctx.statusCode != 0 {
		c.Ctx.Response.WriteHeader(c.Ctx.statusCode)
	}
	c.Ctx.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Ctx.Response.Write([]byte(data))
}
