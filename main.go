package main

import (
	"runtime"
	"snowball/interactor"
	"snowball/xray"

	"github.com/wonderivan/logger"
)

func init() {

	//launch simpler web server
	wf := &xray.WebhookConfig{}
	wf.ReadConfigFile()
	go xray.WebhookServer(wf)

	//logger initialize
	logger.SetLogger("config/logger.json")
	//set max cpus
	runtime.GOMAXPROCS(2)
}

func main() {
	interactor.InterActor()
}
