package webgo
import (
	"net/http"
	"encoding/json"
	"strconv"
	"html/template"
	"bytes"
	"io/ioutil"
)

type (
	Controller struct {
		Ctx *Context
	}
	ControllerInterface interface {
		Init (ctx *Context)
		Prepare ()
		Finish ()
		Error (code int, tpl string)

		GetHeader (key string) string
		SetHeader (key string, val string)

		Redirect(location string, code int)

		Render(tpl_name string, data interface{})
		Json(data interface{})
		Plain(data string)

	}
)

func (c *Controller) Init(ctx *Context) {
	c.Ctx = ctx
}
func (c Controller) Prepare() {}
func (c Controller) Finish() {}
func (c Controller) Error(code int, data string) {
	http.Error(c.Ctx.Response,data,code)
}

func (c Controller) GetHeader(key string)string {
	return c.Ctx.Request.Header.Get(key)
}
func (c Controller) SetHeader(key string, val string) {
	c.Ctx.Response.Header().Set(key, val)
}

func (c Controller) Redirect (location string, code int) {
	http.Redirect(c.Ctx.Response,c.Ctx.Request, location, code)
}

func (c Controller) Render (tpl_name string, data interface{}) {
	var err error
	c.Ctx.Response.Header().Set("Content-Type", "text/html")
	var tpl = template.Must(template.ParseGlob("templates/*"))
	bytes := bytes.NewBufferString("")
	tpl.ExecuteTemplate(bytes, tpl_name+".html", data)

	c.Ctx.body, err = ioutil.ReadAll(bytes)
	if err != nil {
		// TODO: Обработать ошибку
	}
	c.Ctx.Response.Write(c.Ctx.body)
}
func (c Controller) Json (data interface{}) {
	c.Ctx.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	var content []byte
	content, err := json.Marshal(data)
	if err != nil{
		// TODO: Обработать ошибку
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
func (c Controller) Plain (data string) {
	c.Ctx.Response.Header().Set("Content-Type", "text/plain")
	c.Ctx.Response.Write([]byte(data))
}