package webgo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	configFileName = "config"
	CFG_ERROR_TIMEOUT = "errorTimeout"
	CFG_SMTP_HOST = "smtpHost"
	CFG_SMTP_PORT = "smtpPort"
	CFG_SMTP_USER = "smtpUser"
	CFG_SMTP_PASSWORD = "smtpPassword"
	CFG_SMTP_FROM = "smtpFrom"
	CFG_ADMIN_EMAIL = "adminEmail"
)

type config map[string]string

var CFG config

func (cfg *config) Read() (err error) {

	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {

		line = strings.TrimSpace(line)

		if (len(line) < 1) || (line[:1] == ";") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx == -1 {
			sErr := fmt.Sprintf("Invalid config data at line %d: %s", i, line)
			err = errors.New(sErr)
			return
		}

		key := strings.TrimSpace(line[:idx])

		val := strings.TrimSpace(line[idx+1:])

		if len(key) == 0 {
			sErr := fmt.Sprintf("Invalid key at line %d", i)
			err = errors.New(sErr)
			return
		}

		(*cfg)[key] = val
	}

	return
}

func (cfg *config) SetValue(key string, val string) (err error) {

	if len(key) == 0 {
		return
	}

	(*cfg)[key] = val

	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")

	var newLines = make([]string, 0)
	var bFound = false

	for _, line := range lines {

		line = strings.TrimSpace(line)

		if !bFound && (len(line) > 0) && (line[:1] != ";") {

			idx := strings.Index(line, "=")
			if idx >= 0 {

				fKey := strings.TrimSpace(line[:idx])

				if fKey == key {
					line = key + "=" + val
					bFound = true
				}
			}

		}

		newLines = append(newLines, line)

	}

	if !bFound {
		newLines = append(newLines, key+"="+val)
	}

	output := strings.Join(newLines, "\n")
	err = ioutil.WriteFile(configFileName, []byte(output), 0644)

	return
}

func init() {
	CFG = make(config)

	err := CFG.Read()

	if err != nil {
		panic(err)
	}
}