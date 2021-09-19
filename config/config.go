package config

import (
	"io/ioutil"
	"runtime"

	"gopkg.in/yaml.v2"
)

const ConfigFile = "config/config.yaml"
const LockFile = "bin/xray/xray.lock"

var ConfigBuffer []byte
var XrayPath string

func init() {

	ConfigBuffer, _ = ioutil.ReadFile(ConfigFile)
	osType := runtime.GOOS
	if osType == "linux" {
		XrayPath = "bin/xray/xray_linux_amd64"
	} else if osType == "windows" {
		XrayPath = "bin/xray/xray_windows_amd64.exe"
	} else if osType == "darwin" {
		XrayPath = "bin/xray/xray_darwin_amd64"
	}

}

type Configurer interface {
	Get() error
}

//读取配置文件中fofaAPI授权
func GetConfig(conf Configurer) {
	conf.Get()
}

type Auth struct {
	Email string `yaml:"email"`
	Key   string `yaml:"key"`
}

type Fofa struct {
	Auth `yaml:"fofa"`
}

type XrayOptions struct {
	Host    string   `yaml:"host"`
	Port    int      `yaml:"port"`
	Action  string   `yaml:"action"`
	Plugins []string `yaml:"plugins,flow"`
	Pocs    []string `yaml:"pocs,flow"`
}

type Xray struct {
	XrayOptions `yaml:"xray"`
}

type Hook struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

type WebHook struct {
	Hook `yaml:"webhook"`
}

func (fc *Fofa) Get() error {
	yaml.Unmarshal(ConfigBuffer, fc)
	return nil
}

func (xy *Xray) Get() error {
	yaml.Unmarshal(ConfigBuffer, xy)
	return nil
}

func (wh *WebHook) Get() error {
	yaml.Unmarshal(ConfigBuffer, wh)
	return nil
}
