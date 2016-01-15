package webgo
import (
	"net/http"
	"reflect"
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

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var vc reflect.Value
	var Action reflect.Value
	var middlewareGroup string

	method := r.Method
	path := r.URL.Path

	ctx:= Context{Response:w, Request:r}
	if (len(path)>1 && path[len(path)-1:] == "/") {
		http.Redirect(w,r, path[:len(path) - 1], 301)
		return
	}

	// Определем контроллер по прямому вхождению
	if route, ok := a.router.routes[path]; ok {
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
