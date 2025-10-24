package main

import (
	"github.com/saichler/l8logfusion/go/agent/logserver"
	"github.com/saichler/l8utils/go/utils/logger"
)

func main() {
	r := logserver.Main()
	logger.WaitForSignal(r)
}
