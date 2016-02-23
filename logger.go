package webgo

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

// console
// email
// sms
// ???

type ProviderInterface interface {
	GetID() string
	Log(msg []byte)
	Error(msg []byte)
	Fatal(msg []byte)
	Debug(msg []byte)
}

type Logger struct {
	providers      map[string]*ProviderInterface
	logProviders   []string
	errorProviders []string
	fatalProviders []string
	debugProviders []string
}

func NewLogger() *Logger {
	newLogger := Logger{
		providers:    make(map[string]*ProviderInterface, 0),
	}

	return &newLogger
}

func (l *Logger) RegisterProvider(p ProviderInterface) {
	l.providers[p.GetID()] = &p
}

func (l *Logger) AddLogProvider(provIDs ...string) {

	for _, provID := range provIDs {
		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.logProviders {
				if val == pID {
					return
				}
			}

			l.logProviders = append(l.logProviders, pID)
		}
	}
}

func (l *Logger) AddErrorProvider(provIDs ...string) {

	for _, provID := range provIDs {

		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.errorProviders {
				if val == pID {
					return
				}
			}

			l.errorProviders = append(l.errorProviders, pID)
		}
	}
}

func (l *Logger) AddFatalProvider(provIDs ...string) {

	for _, provID := range provIDs {

		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.fatalProviders {
				if val == pID {
					return
				}
			}

			l.fatalProviders = append(l.fatalProviders, pID)
		}
	}
}

func (l *Logger) AddDebugProvider(provIDs ...string) {

	for _, provID := range provIDs {

		p, bFound := l.providers[provID]

		if bFound {
			pID := (*p).GetID()

			for _, val := range l.debugProviders {
				if val == pID {
					return
				}
			}

			l.debugProviders = append(l.debugProviders, pID)
		}
	}
}

func (l *Logger) makeMessage(err []interface{}) []byte {

	t := time.Now()

	buf := bytes.NewBuffer(nil)

	fmt.Fprint(buf, "\n\r\n\r========================== MESSAGE ===========================\n\r")
	fmt.Fprintf(buf, "DateTime: %s\n\r", t.Format(time.ANSIC))

	for i := range err {
		fmt.Fprint(buf, err[i])
		fmt.Fprint(buf, "\n\r")
	}

	return buf.Bytes()
}

func (l *Logger) makeErrorMessage(err []interface{}) []byte {

	t := time.Now()

	trace := make([]byte, 1024)
	runtime.Stack(trace, true)

	buf := bytes.NewBuffer(nil)

	fmt.Fprint(buf, "\n\r\n\r========================== ERROR MESSAGE ===========================\n\r")
	fmt.Fprintf(buf, "DateTime: %s\n\r", t.Format(time.ANSIC))

	for i := range err {
		fmt.Fprint(buf, err[i])
		fmt.Fprint(buf, "\n\r")
	}

	fmt.Fprint(buf, "============================ STACKTRACE ============================\n\r")
	fmt.Fprint(buf, string(trace))
	fmt.Fprint(buf, "\n\r")

	return buf.Bytes()
}

func (l *Logger) Log(err ...interface{}) {
	msg := l.makeMessage(err)

	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Log(msg)
		}
	}
}

func (l *Logger) Error(err ...interface{}) {

	msg := l.makeErrorMessage(err)
	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Error(msg)
		}
	}
}

func (l *Logger) Debug(err ...interface{}) {
	msg := l.makeMessage(err)

	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Debug(msg)
		}
	}
}

func (l *Logger) Fatal(err ...interface{}) {
	msg := l.makeErrorMessage(err)

	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Fatal(msg)
		}
	}

	os.Exit(1)
}

