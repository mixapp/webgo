package webgo

type Definitions struct {
	Handlers map[string][]MiddlewareInterface
}

func (d *Definitions) Register(name string, plugin MiddlewareInterface) {
	if _, ok := d.Handlers[name]; !ok {
		d.Handlers = make(map[string][]MiddlewareInterface)
	}
	d.Handlers[name] = append(d.Handlers[name], plugin)
}
func (m *Definitions) Run(name string, ctx *Context) bool {
	isNext := true

	for _, handler := range m.Handlers[name] {
		if !isNext {
			return false
		}
		isNext = handler.Handler(ctx)
	}

	return isNext
}

type Middleware struct{}

type MiddlewareInterface interface {
	Handler(ctx *Context) bool
}
