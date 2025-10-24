package logs

import (
	"fmt"
	"os"

	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
)

func Main() ifs.IResources {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <vnet ip> <path> <filename> (* for all) \n", os.Args[0])
		return nil
	}
	r := common.NewResources("logs")
	r.SysConfig().RemoteVnet = os.Args[1]
	nic := vnic.NewVirtualNetworkInterface(r, nil)
	lc := &l8logf.L8LogConfig{Path: os.Args[2], Name: os.Args[3]}
	collector := NewLogCollector(lc, nic)
	collector.Collect()
	return nil
}
