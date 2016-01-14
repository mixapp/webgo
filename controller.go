package webgo

type (
	Controller struct {}
	ControllerInterface interface {
		Init()
		Prepare()
		Finish()
	}
)

func (c Controller) Init() {}
func (c Controller) Prepare() {}
func (c Controller) Finish() {}