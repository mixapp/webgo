package webgo

import "fmt"

const (
	PROVIDER_CONSOLE = "console"
	PROVIDER_EMAIL   = "email"
)

type consoleProvider struct {
}

func (p consoleProvider) GetID() string {
	return PROVIDER_CONSOLE
}

func (p consoleProvider) Log(msg []byte) {

	fmt.Println(string(msg))

}

func (p consoleProvider) Error(msg []byte) {

	fmt.Println(string(msg))
}

func (p consoleProvider) Fatal(msg []byte) {

	fmt.Println(string(msg))
}

func (p consoleProvider) Debug(msg []byte) {

	fmt.Println(string(msg))
}

type emailProvider struct {
}

func (p emailProvider) GetID() string {
	return PROVIDER_EMAIL
}

func (p emailProvider) Log(msg []byte) {
	mail := NewMail(CFG[CFG_ADMIN_EMAIL], "Log message", string(msg))
	mail.SendMail()
}

func (p emailProvider) Error(msg []byte) {
	mail := NewMail(CFG[CFG_ADMIN_EMAIL], "Error message", string(msg))
	mail.SendMail()
}

func (p emailProvider) Fatal(msg []byte) {
	mail := NewMail(CFG[CFG_ADMIN_EMAIL], "Fatal message", string(msg))
	mail.SendMail()
}

func (p emailProvider) Debug(msg []byte) {
	mail := NewMail(CFG[CFG_ADMIN_EMAIL], "Debug message", string(msg))
	mail.SendMail()
}
