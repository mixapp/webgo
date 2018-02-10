package webgo

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

type (
	Router struct {
		routes Routes
		once   sync.Once
	}
	Routes map[string]map[string]Route
	Route  struct {
		Keys           []string
		Regex          *regexp.Regexp
		Pattern        string
		ControllerType reflect.Type
		Options        *RouteOptions
	}
	Params map[string]string
	Match  struct {
		Params         Params
		Pattern        string
		ControllerType reflect.Type
		Options        *RouteOptions
	}
	RouteOptions struct {
		MiddlewareGroup string
		Controller      ControllerInterface
		Action          string
		ContentType     string  //Deprecated
		BodyLength      int64
		Timeout         time.Duration
		I18n            bool
	}
)

func (r *Router) Match(method string, url string) (result *Match) {
	r.internalInit()

	routing, ok := r.routes[method]
	if !ok {
		return
	}

	// Определем контроллер по прямому вхождению
	if route, ok := routing[url]; ok {
		result = &Match{make(Params), route.Pattern, route.ControllerType, route.Options}
		return
	}

	for _, route := range routing {
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
		return
	}

	return
}

var (
	_RE_KEY_PATTERN = regexp.MustCompile(`:([A-Za-z0-9]+)`)
	_PATH_REPLACER  = strings.NewReplacer(`/`, `\/`, `.`, `\.`)
)

func (r *Router) Add(method string, path string, opts *RouteOptions) (err error) {
	r.internalInit()
	//c ControllerInterface, action string, middlewareGroup string

	routing, ok := r.routes[method]
	if !ok {
		routing = make(map[string]Route)
		r.routes[method] = routing
	}

	if _, ok := routing[path]; ok {
		err = fmt.Errorf("Route path already use: '%s'->'%s'.", method, path)
		return
	}

	reflectVal := reflect.ValueOf(opts.Controller)
	val := reflectVal.MethodByName(opts.Action)

	if !val.IsValid() {
		err = fmt.Errorf("Not found action for '%s'->'%s': '%s'.", method, path, opts.Action)
		return
	}

	controller := reflect.Indirect(reflectVal).Type()

	/* Добавляем роутер */
	matches := _RE_KEY_PATTERN.FindAllStringSubmatch(path, -1)
	keys := make([]string, len(matches))
	for i := range matches {
		keys[i] = matches[i][1]
	}

	str := fmt.Sprintf("^%s\\/?$", _PATH_REPLACER.Replace(path))
	str = _RE_KEY_PATTERN.ReplaceAllString(str, "([^\\/]+)")

	regex, err := regexp.Compile(str)
	if err != nil {
		return err
	}

	routing[path] = Route{keys, regex, path, controller, opts}
	return
}

func (r *Router) Copy(dest *Router) (err error) {
	r.internalInit()

	for method, srcRouting := range r.routes {
		for routePath, routeDesc := range srcRouting {
			err = dest.Add(method, routePath, routeDesc.Options)
			if err != nil {
				return
			}
		}
	}

	return
}

func (r *Router) internalInit() {
	r.once.Do(func() {
		if r.routes == nil {
			r.routes = make(Routes)
		}
	})
}
