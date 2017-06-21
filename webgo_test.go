package webgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/IntelliQru/config"
)

var TestControllerFunc func(controller *TestController)

type TestController struct {
	Controller
}

func (t *TestController) Invoke() {
	TestControllerFunc(t)
}

func TestMultyPartForm(t *testing.T) {

	// PREPARE ENVIRONMENT

	createTestConfig(t)

	Post("/post", RouteOptions{
		Controller:  new(TestController),
		Action:      "Invoke",
		ContentType: CT_MULTIPART,
	})

	go Run()

	// TEST

	reqParams := map[string]interface{}{
		"file1":  []byte("text of file #1"),
		"file2":  []byte("text of file #2"),
		"field1": "text of field1",
		"field2": "text of field2",
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	TestControllerFunc = func(controller *TestController) {
		defer wg.Done()

		if len(controller.Ctx.Files) != 2 {
			t.Fatal(controller.Ctx.Files)
		} else {
			for _, file := range controller.Ctx.Files {
				if !strings.HasPrefix(file.Name, "file") {
					t.Errorf("Wrong file name: '%s'", file.Name)
				}

				var fileData []byte
				if data, err := ioutil.ReadFile(file.Path); err != nil {
					t.Errorf("Failed read file data '%s': %s", file.Name, err)
				} else {
					fileData = data
				}

				if srcData, ok := reqParams[file.Name]; !ok {
					t.Errorf("Not found source file data: '%s'", file.Name, string(file.Name))
				} else if !bytes.Equal(srcData.([]byte), fileData) {
					t.Errorf("Wrong file data '%s': [%q]!=[%q] ", file.Name, srcData, fileData)
				}
			}
		}

		if len(controller.Ctx.Body) != 2 {
			t.Fatal(controller.Ctx.Body)
		} else {
			for name, data := range controller.Ctx.Body {
				if !strings.HasPrefix(name, "field") {
					t.Errorf("Wrong field name: '%s'", name)
				}

				array, ok := data.([]string)
				if !ok {
					t.Errorf("Wrong field type %s (%T)", name, data)
				}
				if len(array) != 1 {
					t.Errorf("Wrong field data: '%s'", array)
				}

				if srcData, ok := reqParams[name]; !ok {
					t.Errorf("Not found source field data: '%s'", name)
				} else if srcData.(string) != array[0] {
					t.Errorf("Wrong file data '%s': '%s' != '%s' ", name, srcData.(string), array[0])
				}
			}
		}
	}

	{
		// SEND TEST REQUEST
		contentType, reqBody, err := newMultipartForm(reqParams)
		if err != nil {
			wg.Done()
			t.Fatal(err)
		}

		reqHeader := http.Header{}
		reqHeader.Set("Content-Type", contentType)

		statusCode, resHeader, body := sendRequest(t, "post", reqHeader, reqBody)
		if statusCode != 200 {
			wg.Done()
			t.Error(statusCode, resHeader, string(body))
		} else if len(body) != 0 {
			wg.Done()
			t.Error(resHeader, string(body))
		}
	}

	wg.Wait()
}

func sendRequest(t *testing.T, uri string, header http.Header, body []byte) (int, http.Header, []byte) {

	var (
		err error
		req *http.Request
		url = fmt.Sprintf("http://%s:%d/"+uri, CFG.Str("host"), CFG.Int("port"))
	)

	if body != nil {
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		t.Fatal(err)
	}

	req.Close = true
	if body != nil {
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}

	if header != nil {
		for k := range header {
			req.Header.Set(k, header.Get(k))
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return res.StatusCode, res.Header, resBody
}

func newMultipartForm(params map[string]interface{}) (contentType string, content []byte, err error) {

	retval := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(retval)

	for key, val := range params {

		switch val.(type) {
		case []byte:

			var part io.Writer
			part, err = writer.CreateFormFile(key, key)
			if err != nil {
				return
			}

			if _, err = io.Copy(part, bytes.NewReader(val.([]byte))); err != nil {
				return
			}
		case string:
			if err = writer.WriteField(key, val.(string)); err != nil {
				writer.Close()
				return "", nil, err
			}

		default:

			var jd []byte
			jd, err = json.Marshal(val)
			if err != nil {
				return
			}

			if err = writer.WriteField(key, string(jd)); err != nil {
				writer.Close()
				return "", nil, err
			}
		}

	}

	if err = writer.Close(); err != nil {
		return
	}

	contentType = writer.FormDataContentType()
	content = retval.Bytes()

	return
}

func createTestConfig(t *testing.T) {

	fileData := struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}{
		Host: "127.0.1.1",
		Port: 8000,
	}

	jd, err := json.Marshal(fileData)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile("config.json", jd, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	if err = cfg.ReadConfig(); err != nil {
		t.Fatal(err)
	}

	CFG = cfg
}

func removeConfig(t *testing.T) {
	if err := os.Remove("config.json"); err != nil {
		t.Error(err)
	}
}
