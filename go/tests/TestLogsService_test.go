package tests

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8bus/go/overlay/vnet"
	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/agent/logs"
	"github.com/saichler/l8logfusion/go/agent/logserver"
	"github.com/saichler/l8logfusion/go/agent/ui/websvr"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8reflect/go/reflect/introspecting"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils"
	"github.com/saichler/l8utils/go/utils/logger"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	vnetPort = uint16(12443)
)

func startVnet() *vnet.VNet {
	r := utils.NewResources("logsVnet", vnetPort, 0)
	vnt := vnet.NewVNet(r)
	vnt.Start()
	logserver.ActivateLogService(vnt.VnetVnic())
	return vnt
}

func startNic(logDir, logFile string) ifs.IVNic {
	os.MkdirAll(logDir, 0755)
	r := utils.NewResources("logs", vnetPort, 0)
	r.SysConfig().RemoteVnet = protocol.MachineIP
	nic := vnic.NewVirtualNetworkInterface(r, nil)
	nic.Start()
	nic.WaitForConnection()
	fmt.Println("Starting test")
	lc := &l8logf.L8LogConfig{Path: logDir, Name: logFile}
	collector := logs.NewLogCollector(lc, nic)
	go collector.Collect()
	return nic
}

func TestLogsService(t *testing.T) {
	ifs.LogToFiles = false
	logDir := "./logs"
	logFile := "log.log"
	os.MkdirAll(logDir, 0755)
	vnt := startVnet()
	nic := startNic(logDir, logFile)

	filename := "./logs/log.log"
	ll := logger.NewLoggerDirectImpl(logger.NewFileLogMethod(filename))
	defer func() {
		s, _ := os.Stat(filename)
		fmt.Println(s.Size())
		os.Remove(filename)
		os.RemoveAll("./logs")
	}()

	for i := 0; i < 1000; i++ {
		str := "Hello World " + strconv.Itoa(i) + "!"
		ll.Info(str)
	}
	time.Sleep(time.Second * 5)

	ls := logserver.LogsService(vnt.Resources())
	lsFileName := ""
	for k, _ := range ls.Files() {
		lsFileName = k
		break
	}

	lsf, err := os.Stat(lsFileName)
	if err != nil {
		nic.Resources().Logger().Fail(t, "server logs file not exist")
		return
	}

	sf, err := os.Stat(filename)
	if err != nil {
		nic.Resources().Logger().Fail(t, "local logs file not exist")
		return
	}

	if lsf.Size() != sf.Size() {
		nic.Resources().Logger().Fail(t, "local logs file size not equal")
		return
	}

	node, _ := nic.Resources().Introspector().Inspect(l8logf.L8File{})
	introspecting.AddPrimaryKeyDecorator(node, "Path", "Name")

	elems, e := object.NewQuery("select * from l8file where path=/data/logdb/192.168.86.220/logs and name = log.log limit 100 page 0", nic.Resources())
	if e != nil {
		nic.Resources().Logger().Fail(t, "query logs failed: ", e)
		return
	}

	q, _ := elems.Query(nic.Resources())
	jsn, _ := protojson.Marshal(elems.PQuery())
	fmt.Println(string(jsn))
	fmt.Println(q.ValueForParameter("path"))

	gsql := "select * from l8file where path=/data/logdb/192.168.86.220/logs and name = log.log limit 1 page 0"
	resp := nic.Request(nic.Resources().SysConfig().RemoteUuid, common.LogServiceName, common.LogServiceArea, ifs.GET, gsql, 5)
	jsn, _ = protojson.Marshal(resp.Element().(proto.Message))
	os.WriteFile("resp.json", jsn, 0777)

	websvr.StartWebServer(1443, "./test")
}
