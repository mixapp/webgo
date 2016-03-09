package webgo

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type (
	Router struct {
		routes Routes
	}
	Routes map[string]map[string]Route
	Route  struct {
		Keys            []string
		Regex           *regexp.Regexp
		Pattern         string
		ControllerType      reflect.Type
		Options *RouteOptions

	}
	Params map[string]string
	Match  struct {
		Params          Params
		Pattern         string
		ControllerType      reflect.Type
		Options *RouteOptions
	}
	RouteOptions struct {
		MiddlewareGroup string
		Controller ControllerInterface
		Action string
		ContentType string
		BodyLength int
	}
)

func (r *Router) Match(method string, url string) *Match {
	var result *Match

	// Определем контроллер по прямому вхождению
	if route, ok := r.routes[method][url]; ok {
		result = &Match{make(Params), route.Pattern, route.ControllerType, route.Options}
		return result
	}

	for _, route := range r.routes[method] {

		if !route.Regex.MatchString(url) {
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
		result = &Match{params, route.Pattern, route.ControllerType, route.Options}
	}

	return result
}
func (r *Router) addRoute(method string, path string, opts *RouteOptions){
	//c ControllerInterface, action string, middlewareGroup string

	reflectVal := reflect.ValueOf(opts.Controller)
	val := reflectVal.MethodByName(opts.Action)

	if !val.IsValid() {
		LOGGER.Fatal("Action not found: "+opts.Action)
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

	if r.routes[method] == nil {
		r.routes[method] = make(map[string]Route)
	}

	r.routes[method][path] = Route{keys, regex, path, controller, opts}

}
