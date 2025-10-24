package tests

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/agent/logs"
	"github.com/saichler/l8logfusion/go/agent/logserver"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8utils/go/utils/logger"
)

func TestLogsService(t *testing.T) {
	os.MkdirAll("./logs", 0755)
	vr := logserver.Main()

	r := common.NewResources("logs")
	r.SysConfig().RemoteVnet = protocol.MachineIP
	nic := vnic.NewVirtualNetworkInterface(r, nil)
	nic.Start()
	nic.WaitForConnection()
	fmt.Println("Starting test")
	lc := &l8logf.L8LogConfig{Path: "./logs", Name: "log.log"}
	collector := logs.NewLogCollector(lc, nic)
	go collector.Collect()

	filename := "./logs/log.log"
	ll := logger.NewLoggerDirectImpl(logger.NewFileLogMethod(filename))
	defer func() {
		s, _ := os.Stat(filename)
		fmt.Println(s.Size())
		os.Remove(filename)
		os.RemoveAll("./logs")
	}()

	//time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		str := "Hello World " + strconv.Itoa(i) + "!"
		ll.Info(str)
		time.Sleep(time.Millisecond * 110)
	}
	time.Sleep(time.Second)

	ls := logserver.LogsService(vr)
	lsFileName := ""
	for k, _ := range ls.Files() {
		lsFileName = k
		break
	}

	lsf, err := os.Stat(lsFileName)
	if err != nil {
		r.Logger().Fail(t, "server logs file not exist")
		return
	}
	sf, err := os.Stat(filename)
	if err != nil {
		r.Logger().Fail(t, "local logs file not exist")
		return
	}
	if lsf.Size() != sf.Size() {
		r.Logger().Fail(t, "local logs file size not equal")
		return
	}
}
