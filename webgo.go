package webgo

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	//"sync"
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
var LOGGER *Logger

func init() {

	// Init LOGGER
	LOGGER = NewLogger()

	cp := consoleProvider{}
	ep := emailProvider{}

	LOGGER.RegisterProvider(cp)
	LOGGER.RegisterProvider(ep)

	LOGGER.AddLogProvider(PROVIDER_CONSOLE)
	LOGGER.AddErrorProvider(PROVIDER_CONSOLE, PROVIDER_EMAIL)
	LOGGER.AddFatalProvider(PROVIDER_CONSOLE, PROVIDER_EMAIL)
	LOGGER.AddDebugProvider(PROVIDER_CONSOLE)

	// Init App
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

	app.workDir, _ = os.Getwd()
	app.tmpDir = app.workDir + "/tmp"
	app.maxBodyLength = 131072

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
	//cn, ok := w.(http.CloseNotifier)
	//if !ok {
	//	LOGGER.Fatal("don't support CloseNotifier")
	//}

	var vc reflect.Value
	var Action reflect.Value
	var middlewareGroup string

	method := r.Method
	path := r.URL.Path

	// TODO как отдавать статику?
	/*// Отдаем статику если был запрошен файл
	ext := filepath.Ext(path)
	if ext != "" {
		http.ServeFile(w, r, app.staticDir+filepath.Clean(path))
		return
	}*/

	if len(path) > 1 && path[len(path)-1:] == "/" {
		http.Redirect(w, r, path[:len(path)-1], 301)
		return
	}

	route := a.router.Match(method, path)
	if route == nil {
		http.Error(w, "", 404)
		return
	}

	if route.Options.Timeout == 0 {
		route.Options.Timeout = 2
	}
	//timeout := time.After(route.Options.Timeout * time.Second)
	//done := make(chan bool)

	vc = reflect.New(route.ControllerType)
	Action = vc.MethodByName(route.Options.Action)
	middlewareGroup = route.Options.MiddlewareGroup

	var err error
	ctx := Context{Action: route.Options.Action, Response: w, Request: r, Query: make(map[string]interface{}), Body: make(map[string]interface{}), Params: route.Params, Method: method}
	ctx.ContentType = ctx.Request.Header.Get("Content-Type")
	ctx.ContentType, _, err = mime.ParseMediaType(ctx.ContentType)

	if err != nil && method != "GET" {
		http.Error(w, "", 400)
		return
	}

	if route.Options.ContentType != "" && (method == "POST" || method == "PUT") {
		if route.Options.ContentType != ctx.ContentType {
			http.Error(w, "", 400)
			return
		}
	}

	Controller, ok := vc.Interface().(ControllerInterface)
	if !ok {
		LOGGER.Error(errors.New("controller is not ControllerInterface"))
		http.Error(w, "", 500)
		return
	}

	// Парсим запрос
	var maxBodyLength int64 = app.maxBodyLength
	if route.Options.BodyLength > 0 {
		maxBodyLength = route.Options.BodyLength
	}

	code, err := parseRequest(&ctx, maxBodyLength)
	if err != nil {
		http.Error(w, "", code)
		return
	}

	// Инициализация контекста
	Controller.Init(&ctx)

	// Запуск предобработчика
	if !Controller.Prepare() {
		Controller.exec()
		return
	}

	// Запуск цепочки middleware
	if middlewareGroup != "" {
		isNext := app.definitions.Run(middlewareGroup, &ctx)
		if !isNext {
			return
		}
	}

	in := make([]reflect.Value, 0)
	Action.Call(in)
	//go func () {
	//	in := make([]reflect.Value, 0)
	//	Action.Call(in)
	//	done <- true
	//}()

	// Запуск постобработчика

	Controller.Finish()

	if ctx.ContentType == "multipart/form-data" {
		err = ctx.Files.RemoveAll()
		if err != nil {
			LOGGER.Error(err)
		}

		err = ctx.Request.MultipartForm.RemoveAll()
		if err != nil {
			LOGGER.Error(err)
		}
	}

	Controller.exec()

	//select {
	//case <-timeout:
	//	ctx.close = true
	//	w.WriteHeader(503)
	//	w.Write([]byte(""))
	//	return
	//case <-cn.CloseNotify():
	//	//TODO: НИХРЕНА НЕПОНЯТНО!!!
	//	ctx.close = true
	//	w.WriteHeader(503)
	//	w.Write([]byte(""))
	//	return
	//case <-done:
	//	// TODO: Обработать ошибки
	//	if ctx.error != nil {
	//		if ctx.code == 0 {
	//			ctx.code = 500
	//		}
	//		ctx.Response.WriteHeader(ctx.code)
	//		ctx.Response.Write(ctx.output)
	//		return
	//	}
	//
	//	// Проверяем редирект
	//	if ctx.IsRedirect(){
	//		ctx.Response.WriteHeader(ctx.code)
	//		return
	//	}
	//
	//	// Выводим данные
	//	if ctx.code == 0 {
	//		ctx.code = 200
	//	}
	//	ctx.Response.WriteHeader(ctx.code)
	//	ctx.Response.Write(ctx.output)
	//	return
	//}

}

func RegisterMiddleware(name string, plugins ...MiddlewareInterface) {
	for _, plugin := range plugins {
		app.definitions.Register(name, plugin)
	}
}
func RegisterModule(name string, module ModuleInterface) {
	app.modules.RegisterModule(name, module)
}
func Get(url string, opts RouteOptions) {
	app.router.addRoute("GET", url, &opts)
}
func Post(url string, opts RouteOptions) {
	app.router.addRoute("POST", url, &opts)
}
func Put(url string, opts RouteOptions) {
	app.router.addRoute("PUT", url, &opts)
}
func Delete(url string, opts RouteOptions) {
	app.router.addRoute("DELETE", url, &opts)
}
func Options(url string, opts RouteOptions) {
	app.router.addRoute("OPTIONS", url, &opts)
}

func GetModule(str string) ModuleInterface {
	return app.modules[str]
}

func Run() {
	var r *int = flag.Int("r", 0, "read timeout")
	var w *int = flag.Int("w", 0, "write timeout")

	port := CFG.Int("port")

	if port == 0 {
		port = 80
	}

	host := CFG.Str("host")
	if host == "" {
		host = "127.0.0.1"
	}

	address := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("WebGO start ", address)

	server := http.Server{
		Addr:         address,
		ReadTimeout:  time.Duration(*r) * time.Second,
		WriteTimeout: time.Duration(*w) * time.Second,
		Handler:      &app,
	}

	//server.SetKeepAlivesEnabled(false)

	err := server.ListenAndServe()
	if err != nil {
		LOGGER.Fatal(err)
	}

}
