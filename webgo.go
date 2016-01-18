package webgo
import (
	"net/http"
	"reflect"
	"net/url"
	"errors"
	"io/ioutil"
	"io"
	"encoding/json"
	"html/template"
	"path/filepath"
	"os"
	"strings"
	"fmt"
	"mime"
)

type App struct {
	router Router
	definitions Definitions
	templates *template.Template
	staticDir string
	modules Modules
}

var app App

func init(){
	templates := template.New("template")
	filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			templates.ParseFiles(path)
		}
		return nil
	})
	app = App{}
	app.router = Router{make(Routes)}
	app.definitions = Definitions{}
	app.templates = templates
	app.staticDir = "public"
	app.modules = Modules{}
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


	switch ctx.ContentType {
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
		g:=ctx.Request.ParseForm()
		fmt.Println("",ctx.Request.PostForm,ctx.Request.Form,g)

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
	if (ctx.Request.Method == "GET") {
		err = ctx.Request.ParseForm()
		// TODO: скопировать данные
		return
	}


	if ctx.Request.Method != "POST" && ctx.Request.Method != "PUT" && ctx.Request.Method != "PATCH" {
		return
	}

	ctx.ContentType = ctx.Request.Header.Get("Content-Type")
	ctx.ContentType, _, err = mime.ParseMediaType(ctx.ContentType)

	if err != nil {
		http.Error(ctx.Response, "", 400)
		return
	}

	if ctx.ContentType != "application/json" &&
		ctx.ContentType != "application/x-www-form-urlencoded" &&
		ctx.ContentType != "multipart/form-data" {
			err = errors.New("Bad Request")
			http.Error(ctx.Response, "", 400)
			return
	}

	// TODO: Правильно спарсить + скопировать данные
	err = parseBody(ctx)
	return
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var vc reflect.Value
	var Action reflect.Value
	var middlewareGroup string

	method := r.Method
	path := r.URL.Path

	// Отдаем статику если был запрошен файл
	ext:= filepath.Ext(path)
	if ext != "" {
		http.ServeFile(w, r, app.staticDir+filepath.Clean(path))
		return
	}

	if (len(path)>1 && path[len(path)-1:] == "/") {
		http.Redirect(w,r, path[:len(path) - 1], 301)
		return
	}

	// Определем контроллер по прямому вхождению
	if route, ok := a.router.routes[path]; ok {
		if route.Method != method {
			http.Error(w, "", 404)
			return
		}

		vc = reflect.New(route.Controller)
		Action = vc.MethodByName(route.Action)
		middlewareGroup = route.MiddlewareGroup
	} else {
		// Определяем контроллер по совпадениям
		route := a.router.Match(method,path)
		if route == nil{
			http.Error(w, "", 404)
			return
		} else {
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

	ctx:= Context{Response:w, Request:r, Query: make(map[string]interface{}), Body: make(map[string]interface{}), Method:method}

	// Парсим запрос
	err := parseRequest(&ctx)
	if err != nil {
		return
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

	// Обрабатываем ошибки
	if ctx.error != nil {
		// TODO: Записать в лог
		http.Error(w, "", 500)
		return
	}

	// Запуск постобработчика
	Controller.Finish()
}

func RegisterMiddleware(name string, plugins ...MiddlewareInterface)  {
	for _, plugin:= range plugins {
		app.definitions.Register(name, plugin)
	}
}
func RegisterModule(name string, module ModuleInterface)  {
	app.modules.RegisterModule(name, module)
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

func MODULES(str string) ModuleInterface {
	return app.modules[str]
}
func Run()  {
	if CFG["port"] == "" {
		LOGGER.Fatal("Unknow port")
	}
	http.ListenAndServe(":"+CFG["port"], &app)
}