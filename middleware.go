package webgo

type Definitions struct {
	Handlers map[string][]MiddlewareInterface
}

func (d *Definitions) Register(name string, plugin MiddlewareInterface) {

	if d.Handlers == nil {
		d.Handlers = make(map[string][]MiddlewareInterface)
	}
	if _, ok := d.Handlers[name]; !ok {
		d.Handlers[name] = make([]MiddlewareInterface, 0)
	}
	d.Handlers[name] = append(d.Handlers[name], plugin)
}
func (m *Definitions) Run(name string, ctx *Context) (bool) {
	if len(name) == 0 {
		return true
	}

	defs, ok := m.Handlers[name]
	if !ok {
		return false
	}

	for _, handler := range defs {
		isNext := handler.Handler(ctx)
		if !isNext {
			return false
		}
	}

	return true
}

type Middleware struct{}

type MiddlewareInterface interface {
	Handler(ctx *Context) bool
}
