package webgo

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
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

const DEFAULT_ERROR_TIMEOUT = 60

type Logger struct {
	timeout        int
	providers      map[string]*ProviderInterface
	logProviders   []string
	errorProviders []string
	fatalProviders []string
	debugProviders []string
	errorsBuffer   map[int]time.Time
}

func NewLogger() *Logger {
	newLogger := Logger{
		providers:    make(map[string]*ProviderInterface, 0),
		timeout:      DEFAULT_ERROR_TIMEOUT,
		errorsBuffer: make(map[int]time.Time, 0),
	}

	errTimeout, bFound := CFG[CFG_ERROR_TIMEOUT]
	if bFound {
		iTimeout, err := strconv.Atoi(errTimeout)
		if err == nil {
			newLogger.timeout = iTimeout
		}
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

			l.errorProviders = append(l.debugProviders, pID)
		}
	}
}

func (l *Logger) makeMessage(err ...interface{}) []byte {

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

func (l *Logger) makeErrorMessage(erorCode int, err ...interface{}) []byte {

	t := time.Now()

	trace := make([]byte, 1024)
	runtime.Stack(trace, true)

	buf := bytes.NewBuffer(nil)

	fmt.Fprint(buf, "\n\r\n\r========================== ERROR MESSAGE ===========================\n\r")
	fmt.Fprintf(buf, "Error code: %d\n\r", erorCode)
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

func (l *Logger) Error(errorCode int, err ...interface{}) {

	t, bFound := l.errorsBuffer[errorCode]

	if bFound {
		d := int(time.Since(t).Seconds())
		if d < l.timeout {
			return
		}
	}

	l.errorsBuffer[errorCode] = time.Now()

	error := l.makeErrorMessage(errorCode, err)
	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Error(error)
		}
	}
}

func (l *Logger) Debug(err ...interface{}) {
	msg := l.makeMessage(err)

	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Log(msg)
		}
	}
}

func (l *Logger) Fatal(errorCode int, err ...interface{}) {
	msg := l.makeErrorMessage(errorCode, err)

	for _, pID := range l.logProviders {
		p, bFound := l.providers[pID]
		if bFound {
			(*p).Fatal(msg)
		}
	}

	os.Exit(1)
}

var LOGGER *Logger

func init() {
	LOGGER = NewLogger()

	cp := consoleProvider{}
	ep := emailProvider{}

	LOGGER.RegisterProvider(cp)
	LOGGER.RegisterProvider(ep)

	LOGGER.AddLogProvider(PROVIDER_CONSOLE)
	LOGGER.AddErrorProvider(PROVIDER_CONSOLE, PROVIDER_EMAIL)
	LOGGER.AddFatalProvider(PROVIDER_CONSOLE, PROVIDER_EMAIL)
	LOGGER.AddDebugProvider(PROVIDER_CONSOLE)
}
