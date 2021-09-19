package main

import (
	"runtime"
	"snowball/interactor"
	"snowball/xray"

	"github.com/wonderivan/logger"
)

func init() {
	go xray.WebhookServer()
	logger.SetLogger("config/logger.json")
	runtime.GOMAXPROCS(2)
}

func main() {
	interactor.InterActor()
}
