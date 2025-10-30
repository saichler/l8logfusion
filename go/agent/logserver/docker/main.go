package main

import (
	"os"
	"strconv"

	"github.com/saichler/l8bus/go/overlay/vnet"
	"github.com/saichler/l8logfusion/go/agent/logserver"
	"github.com/saichler/l8utils/go/utils"
	"github.com/saichler/l8utils/go/utils/logger"
)

func main() {
	vnetPort, _ := strconv.Atoi(os.Args[1])
	r := utils.NewResources("logsVnet", uint16(vnetPort), 0)
	vnt := vnet.NewVNet(r)
	vnt.Start()
	logserver.ActivateLogService(vnt.VnetVnic())
	logger.WaitForSignal(r)
}
