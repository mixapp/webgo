package webgo

var MODULES Modules

type (
	Modules struct {}
	Module struct {}
	ModuleInterface interface {
		Init ()
		ReInit ()
	}
)


func init() {
	MODULES = Modules{}
}