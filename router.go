package webgo
import (
	"regexp"
	"fmt"
	"strings"
	"reflect"
)

type (
	Router struct {
		routes Routes
	}
	Routes map[string]Route
	Route struct {
		Keys []string
		Regex *regexp.Regexp
		Pattern string
		Controller reflect.Type
		Action string
		Method string
		MiddlewareGroup string
	}
	Params map[string]string
	Match struct {
		Params Params
		Pattern string
		Controller reflect.Type
		Action string
		Method string
		MiddlewareGroup string
	}
)
func (r *Router) Match (method string, url string) *Match {
	var result *Match

	for _, route := range r.routes {
		if !route.Regex.MatchString(url) || route.Method != method {
			continue
		}

		match := route.Regex.FindAllStringSubmatch(url, -1)[0][1:]
		params := make(Params)

		for i := range match {
			if len(route.Keys) <= i {
				break
			}
			params[route.Keys[i]] = match[i]
		}
		result = &Match{params, route.Pattern, route.Controller,route.Action, method, route.MiddlewareGroup}
	}

	return result
}
func (r *Router) addRoute(method string, path string, c ControllerInterface, action string, middlewareGroup string) {
	reflectVal := reflect.ValueOf(c)
	val := reflectVal.MethodByName(action);

	if !val.IsValid(){
		// TODO: Заменить панику
		panic("Экшен не найден");
	}

	controller := reflect.Indirect(reflectVal).Type()

	/* Добавляем роутер */
	pattern, _ := regexp.Compile(":([A-Za-z0-9]+)")
	matches := pattern.FindAllStringSubmatch(path, -1)
	keys := []string{}

	for i := range matches {
		keys = append(keys, matches[i][1])
	}

	str := fmt.Sprintf("^%s\\/?$", strings.Replace(path, "/", "\\/", -1))
	str = pattern.ReplaceAllString(str, "([^\\/]+)")
	str = strings.Replace(str, ".", "\\.", -1)

	regex, _ := regexp.Compile(str)
	r.routes[path] = Route{keys, regex, path, controller, action, method, middlewareGroup}

}

