package webgo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
)

const (
	configFileName    = "config"
	CFG_ERROR_TIMEOUT = "errorTimeout"
	CFG_SMTP_HOST     = "smtpHost"
	CFG_SMTP_PORT     = "smtpPort"
	CFG_SMTP_USER     = "smtpUser"
	CFG_SMTP_PASSWORD = "smtpPassword"
	CFG_SMTP_FROM     = "smtpFrom"
	CFG_ADMIN_EMAIL   = "adminEmail"
)

var SALT string = "SsUper!Se@cretKeyBPM2_0Ugsk&EEdh" //32 bytes

var CFG Config

type Config struct {
	data 	map[string]interface{}
	workDir	string
}

func init() {

	CFG = Config{
		data: make(map[string]interface{}),
	}

	err := CFG.ReadConfig()
	if err != nil {
		panic(err)
	}
}

func (cfg *Config) Str(key string) (res string) {

	res, _ = cfg.data[key].(string)
	return
}

func (cfg *Config) Bool(key string) (res bool) {

	res, _ = cfg.data[key].(bool)
	return
}

func (cfg *Config) Int(key string) (res int) {

	res, _ = cfg.data[key].(int)
	return
}

func (cfg *Config) Float64(key string) (res float64) {

	res, _ = cfg.data[key].(float64)
	return
}

func (cfg *Config) Array(key string) (res []interface{}) {

	res, _ = cfg.data[key].([]interface{})

	return
}

func (cfg *Config) ArrayStr(key string) (res []string) {

	res = make([]string, 0)
	arr, ok := cfg.data[key].([]interface{})
	if ok {
		for _, val := range arr {
			str, ok := val.(string)
			if ok {
				res = append(res, str)
			}
		}
	}

	return
}

func (cfg *Config) Map(key string) (res map[string]interface{}) {

	res, _ = cfg.data[key].(map[string]interface{})

	return
}

func (cfg *Config) MapStr(key string) (res map[string]string) {

	res = make(map[string]string, 0)
	m, ok := cfg.data[key].(map[string]interface{})
	if ok {
		for key, val := range m {
			strVal, ok := val.(string)
			if ok {
				m[key] = strVal
			}
		}
	}

	return
}

func (cfg *Config) ReadConfig() (err error) {
	var encConfig string
	cfg.workDir, err = os.Getwd()
	if err != nil {
		return
	}

	configData, err := ioutil.ReadFile(cfg.workDir + "/config.json")
	if err != nil {
		return
	}

	err = json.Unmarshal(configData, &cfg.data)
	if err != nil {
		configData, err = cfg.decrypt([]byte(SALT), configData)
		if err != nil {
			return
		}
		err = json.Unmarshal(configData, &cfg.data)
		if err != nil {
			return
		}
	} else {
		encConfig, err = cfg.encrypt([]byte(SALT), []byte(configData))
		if err != nil {
			return
		}
		err = ioutil.WriteFile(cfg.workDir+"/config.json", []byte(encConfig), 644)
		if err != nil {
			return
		}
	}

	return
}

func (cfg *Config) encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func (cfg *Config) decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func (cfg *Config) encrypt(key, text []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return cfg.encodeBase64(ciphertext), nil
}

func (cfg *Config) decrypt(key, text []byte) ([]byte, error) {
	text = cfg.decodeBase64(string(text))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}