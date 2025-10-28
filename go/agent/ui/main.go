package main

import (
	"github.com/saichler/l8logfusion/go/agent/ui/websvr"
	"github.com/saichler/l8types/go/ifs"
)

func main() {
	ifs.LogToFiles = true
	websvr.StartWebServer(12443, "/data/probler")
}
