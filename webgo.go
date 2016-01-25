package webgo

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type App struct {
	router        Router
	definitions   Definitions
	templates     *template.Template
	staticDir     string
	modules       Modules
	workDir       string
	tmpDir        string
	maxBodyLength int64
}

const (
	CT_JSON      = "application/json"
	CT_FORM      = "application/x-www-form-urlencoded"
	CT_MULTIPART = "multipart/form-data"
)

var app App

func init() {
	var err error
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

	app.workDir, err = os.Getwd()
	app.tmpDir = app.workDir + "/tmp"

	if CFG["maxBodyLength"] == "" {
		panic("maxBodyLength is empty")
	}
	app.maxBodyLength, err = strconv.ParseInt(CFG["maxBodyLength"], 10, 64)
	if err != nil {
		os.Exit(1)
	}

	//TODO: Проверить папку tmp, создать если необходимо
}

func parseRequest(ctx *Context, limit int64) (errorCode int, err error) {
	var body []byte

	defer func() {
		r := recover()
		if r != nil {
			errorCode = 400
			err = errors.New("Bad Request")
		}
	}()
	ctx.Request.Body = http.MaxBytesReader(ctx.Response, ctx.Request.Body, limit)

	if ctx.Request.Method == "GET" {
		err = ctx.Request.ParseForm()
		if err != nil {
			errorCode = 400
			return
		}

		// Копируем данные
		for i := range ctx.Request.Form {
			ctx.Query[i] = ctx.Request.Form[i]
		}

		return
	}

	ctx.ContentType = ctx.Request.Header.Get("Content-Type")
	ctx.ContentType, _, err = mime.ParseMediaType(ctx.ContentType)

	if err != nil {
		errorCode = 400
		return
	}

	switch ctx.ContentType {
	case CT_JSON:
		body, err = ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			errorCode = 400
			return
		}

		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			errorCode = 400
			return
		}
		ctx._Body = body
		ctx.Body = data.(map[string]interface{})

		return
	case CT_FORM:
		err = ctx.Request.ParseForm()
		if err != nil {
			errorCode = 400
			return
		}

	case CT_MULTIPART:
		err = ctx.Request.ParseMultipartForm(limit)
		if err != nil {
			//TODO: 400 or 413
			errorCode = 400
			return
		}

		for _, fheaders := range ctx.Request.MultipartForm.File {
			for _, hdr := range fheaders {
				var infile multipart.File
				if infile, err = hdr.Open(); nil != err {
					errorCode = 500
					return
				}

				var outfile *os.File
				if outfile, err = os.Create(app.tmpDir + "/" + hdr.Filename); nil != err {
					errorCode = 500
					return
				}
				// 32K buffer copy
				var written int64
				if written, err = io.Copy(outfile, infile); nil != err {
					errorCode = 500
					return
				}

				ctx.Files = append(ctx.Files, File{FileName: hdr.Filename, Size: int64(written)})
			}
		}

	default:
		err = errors.New("Bad Request")
		errorCode = 400
		return
	}

	for i := range ctx.Request.Form {
		ctx.Body[i] = ctx.Request.Form[i]
	}

	return
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var vc reflect.Value
	var Action reflect.Value
	var middlewareGroup string
	var Params map[string]string

	method := r.Method
	path := r.URL.Path

	// Отдаем статику если был запрошен файл
	ext := filepath.Ext(path)
	if ext != "" {
		http.ServeFile(w, r, app.staticDir+filepath.Clean(path))
		return
	}

	if len(path) > 1 && path[len(path)-1:] == "/" {
		http.Redirect(w, r, path[:len(path)-1], 301)
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
		route := a.router.Match(method, path)
		if route == nil {
			http.Error(w, "", 404)
			return
		} else {
			vc = reflect.New(route.Controller)
			Action = vc.MethodByName(route.Action)
			middlewareGroup = route.MiddlewareGroup
			Params = route.Params
		}
	}

	Controller, ok := vc.Interface().(ControllerInterface)
	if !ok {
		LOGGER.Error(2,errors.New("controller is not ControllerInterface"))
		http.Error(w, "", 500)
		return
	}

	ctx := Context{Response: w, Request: r, Query: make(map[string]interface{}), Body: make(map[string]interface{}), Params: Params, Method: method}

	// Парсим запрос
	code, err := parseRequest(&ctx, app.maxBodyLength)
	if err != nil {
		http.Error(w, "", code)
		return
	}

	// Инициализация контекста
	Controller.Init(&ctx)

	// Запуск предобработчика
	Controller.Prepare()

	// Запуск цепочки middleware
	if middlewareGroup != "" {
		isNext := app.definitions.Run(middlewareGroup, &ctx)
		if !isNext {
			return
		}
	}

	// Запуск Экшена
	in := make([]reflect.Value, 0)
	Action.Call(in)

	if ctx.ContentType == "multipart/form-data" {
		err = ctx.Files.RemoveAll()
		if err != nil {
			LOGGER.Error(3,err)
		}

		err = ctx.Request.MultipartForm.RemoveAll()
		if err != nil {
			LOGGER.Error(4,err)
		}
	}

	// Обрабатываем ошибки
	if ctx.error != nil {
		LOGGER.Error(5,err)
		http.Error(w, "", 500)
		return
	}

	// Запуск постобработчика
	Controller.Finish()
}

func RegisterMiddleware(name string, plugins ...MiddlewareInterface) {
	for _, plugin := range plugins {
		app.definitions.Register(name, plugin)
	}
}
func RegisterModule(name string, module ModuleInterface) {
	app.modules.RegisterModule(name, module)
}

func Get(url string, controller ControllerInterface, middlewareGroup string, action string) {
	app.router.addRoute("GET", url, controller, action, middlewareGroup)
}
func Post(url string, controller ControllerInterface, middlewareGroup string, action string) {
	app.router.addRoute("POST", url, controller, action, middlewareGroup)
}
func Put(url string, controller ControllerInterface, middlewareGroup string, action string) {
	app.router.addRoute("PUT", url, controller, action, middlewareGroup)
}
func Delete(url string, controller ControllerInterface, middlewareGroup string, action string) {
	app.router.addRoute("DELETE", url, controller, action, middlewareGroup)
}
func Options(url string, controller ControllerInterface, middlewareGroup string, action string) {
	app.router.addRoute("OPTIONS", url, controller, action, middlewareGroup)
}

func MODULES(str string) ModuleInterface {
	return app.modules[str]
}
func Run() {
	if CFG["port"] == "" {
		LOGGER.Fatal(1,"Unknow port")
	}
	http.ListenAndServe(":"+CFG["port"], &app)
}
