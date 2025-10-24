package logs

import (
	"os"

	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
)

func Main() ifs.IResources {
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

	r := common.NewResources("logs")
	r.SysConfig().RemoteVnet = ip
	nic := vnic.NewVirtualNetworkInterface(r, nil)
	nic.Start()
	nic.WaitForConnection()
	
	lc := &l8logf.L8LogConfig{Path: logpath, Name: logfile}
	collector := NewLogCollector(lc, nic)
	collector.Collect()
	return nil
}
