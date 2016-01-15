package webgo
import (
	"net/http"
	"reflect"
	"net/url"
	"errors"
	"io/ioutil"
	"io"
	"encoding/json"
)

type App struct {
	router Router
	definitions Definitions
}

var app App

func init(){
	app = App{}
	app.router = Router{make(Routes)}
	app.definitions = Definitions{}
}

func parseBody(ctx *Context) (err error) {
	var body []byte
	defer func() {
		r:=recover()
		if r != nil {
			http.Error(ctx.Response, "", 400)
			err = errors.New("Bad Request")
		}
	}()

	contentType := ctx.Request.Header.Get("Content-Type")
	if len(contentType) > 33 && contentType[0:33] == "application/x-www-form-urlencoded"{
		contentType = "application/x-www-form-urlencoded"
	}

	ctx.ContentType = contentType

	switch contentType {
		case "application/json":
			body, err = ioutil.ReadAll(ctx.Request.Body)
			if err != nil {
				http.Error(ctx.Response, "", 400)
				return
			}

			var data interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				http.Error(ctx.Response, "", 400)
				return
			}
			ctx._Body = body
			ctx.Body = data.(map[string]interface{})

			return

		case "application/x-www-form-urlencoded":
			// TODO Может быть проблема с чтением пустого запроса EOF
			var reader io.Reader = ctx.Request.Body
			var values url.Values

			maxFormSize := int64(10 << 20)
			reader = io.LimitReader(ctx.Request.Body, maxFormSize+1)

			body, err = ioutil.ReadAll(reader)
			if err != nil {
				http.Error(ctx.Response, "", 400)
				return
			}

			if int64(len(body)) > maxFormSize {
				http.Error(ctx.Response, "", 413)
				err = errors.New("Request Entity Too Large")
				return
			}

			values, err = url.ParseQuery(string(body))

			if err != nil{
				http.Error(ctx.Response, "", 400)
				return
			}

			for i := range values{
				if len(values[i]) == 1{
					ctx.Body[i] = values[i][0]
				} else {
					ctx.Body[i] = values[i]
				}
			}
			ctx._Body = body
			return
		case "multipart/form-data":
			return
		default:
			err = errors.New("Bad Request")
			http.Error(ctx.Response, "", 400)
			return
	}

	return err
}

func parseRequest (ctx *Context) (err error){
	var values url.Values
	if (ctx.Request.Method == "GET") {
		values, err = url.ParseQuery(ctx.Request.URL.RawQuery)
		if err != nil{
			return
		}
		for i := range values{
			if len(values[i]) == 1{
				ctx.Query[i] = values[i][0]
			} else {
				ctx.Query[i] = values[i]
			}
		}
		return
	}

	if (ctx.Request.Method == "POST") {
		err = parseBody(ctx)
		return
	}

	return
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var vc reflect.Value
	var Action reflect.Value
	var middlewareGroup string

	method := r.Method
	path := r.URL.Path

	ctx:= Context{Response:w, Request:r, Query: make(map[string]interface{}), Body: make(map[string]interface{})}

	if (len(path)>1 && path[len(path)-1:] == "/") {
		http.Redirect(w,r, path[:len(path) - 1], 301)
		return
	}

	// Отдаем статику если был запрошен файл
	// TODO: Реализовать отдачу файлов


	// Парсим запрос
	err := parseRequest(&ctx)
	if err != nil {
		return
	}

	// Определем контроллер по прямому вхождению
	if route, ok := a.router.routes[path]; ok {
		// TODO: Добавить проверку метода

		vc = reflect.New(route.Controller)
		Action = vc.MethodByName(route.Action)
		middlewareGroup = route.MiddlewareGroup
	} else {
		// Определяем контроллер по совпадениям
		route := a.router.Match(method,path)
		if route == nil{
			// TODO 404
			http.Error(w, "", 404)
			return;
		} else {
			// TODO: Добавить проверку метода
			vc = reflect.New(route.Controller)
			Action = vc.MethodByName(route.Action)
			middlewareGroup = route.MiddlewareGroup
		}
	}

	Controller, ok := vc.Interface().(ControllerInterface)
	if !ok {
		// TODO: Заменить панику
		panic("controller is not ControllerInterface")
	}

	// Инициализация контекста
	Controller.Init(&ctx)

	// Запуск предобработчика
	Controller.Prepare()

	// Запуск цепочки middleware
	if middlewareGroup != "" {
		isNext := app.definitions.Run(middlewareGroup,&ctx)
		if !isNext {
			return
		}
	}

	// Запуск Экшена
	in := make([]reflect.Value, 0)
	Action.Call(in)

	// Запуск постобработчика
	Controller.Finish()
}

func RegisterMiddleware(name string, plugins ...MiddlewareInterface)  {
	for _, plugin:= range plugins {
		app.definitions.Register(name, plugin)
	}

}

func Get(url string, controller ControllerInterface, middlewareGroup string, flags []string, action string) {
	app.router.addRoute("GET", url, controller, action, middlewareGroup)
}
func Post(url string, controller ControllerInterface, middlewareGroup string, flags []string, action string) {
	app.router.addRoute("POST", url, controller, action, middlewareGroup)
}
func Put(url string, controller ControllerInterface, middlewareGroup string, flags []string, action string) {
	app.router.addRoute("PUT", url, controller, action, middlewareGroup)
}
func Delete(url string, controller ControllerInterface, middlewareGroup string, flags []string, action string) {
	app.router.addRoute("DELETE", url, controller, action, middlewareGroup)
}
func Options(url string, controller ControllerInterface, middlewareGroup string, flags []string, action string)  {
	app.router.addRoute("OPTIONS", url, controller, action, middlewareGroup)
}

func Run(port string)  {
	http.ListenAndServe(port, &app)
}
