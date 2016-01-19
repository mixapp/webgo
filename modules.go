package webgo

type (
	Modules map[string]ModuleInterface
	ModuleInterface interface {
		Init ()
		GetInstance () interface{}
	}
)

func (m Modules) RegisterModule (name string, module ModuleInterface) {
	module.Init()
	app.modules[name] = module
}
