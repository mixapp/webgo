package webgo
import (
	"fmt"
	"runtime"
	"bytes"
	"os"
)

var LOGGER Logger

type (
	errorMessage struct {
		timestamp string
		message string
		code int
		stack []byte
	}
	Logger struct {}
)

func (l *Logger) getMessage(err []interface{}) []byte {
	trace := make([]byte, 1024)
	runtime.Stack(trace, true)

	buf := bytes.NewBuffer(nil)
	fmt.Fprint(buf,"========================== ERROR MESSAGE ===========================\n\r")
	for i:= range err{
		fmt.Fprint(buf,err[i])
		fmt.Fprint(buf,"\n\r")
	}
	fmt.Fprint(buf,"============================ STACKTRACE ============================\n\r")
	fmt.Fprint(buf,string(trace))
	fmt.Fprint(buf,"\n\r")

	return buf.Bytes()
}

func (l *Logger) Log (err...interface{}) {
	error := l.getMessage(err)
	fmt.Println("\n\r"+string(error)+"\n\r\n\r")
}
func (l *Logger) Error (err...interface{}) {
	error := l.getMessage(err)
	fmt.Println("\n\r"+string(error)+"\n\r\n\r")
}
func (l *Logger) Debug (err...interface{}) {
	error := l.getMessage(err)
	fmt.Println("\n\r"+string(error)+"\n\r\n\r")
}
func (l *Logger) Fatal (err...interface{}) {
	error := l.getMessage(err)
	fmt.Println("\n\r"+string(error)+"\n\r\n\r")
	os.Exit(1)
}

func init() {
	LOGGER = Logger{}
}