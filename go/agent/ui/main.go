package main

import (
	"github.com/saichler/l8logfusion/go/agent/ui/websvr"
)

func main() {
	websvr.StartWebServer(26000, "/data/probler")
}
