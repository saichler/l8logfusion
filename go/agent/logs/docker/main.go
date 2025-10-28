package main

import (
	"os"

	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/logs"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8utils/go/utils"
)

func main() {
	ip := os.Getenv("NODE_IP")
	if ip == "" {
		panic("Env variable NODE_IP is not set")
	}

	logpath := os.Getenv("LOGPATH")
	if logpath == "" {
		panic("Env variable LOGPATH is not set")
	}

	logfile := os.Getenv("LOGFILE")
	if logfile == "" {
		panic("Env variable LOGFILE is not set")
	}

	r := utils.NewResources("logs", 12443, 30)
	r.SysConfig().RemoteVnet = ip

	nic := vnic.NewVirtualNetworkInterface(r, nil)
	nic.Start()
	nic.WaitForConnection()

	lc := &l8logf.L8LogConfig{Path: logpath, Name: logfile}
	collector := logs.NewLogCollector(lc, nic)
	collector.Collect()
}
