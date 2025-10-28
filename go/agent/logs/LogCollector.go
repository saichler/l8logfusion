package logs

import (
	"fmt"
	"os"
	"time"

	"github.com/saichler/l8bus/go/overlay/protocol"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/strings"
	common2 "github.com/saichler/netop/go/common"
)

type LogCollector struct {
	logConfig *l8logf.L8LogConfig
	vnic      ifs.IVNic
}

func NewLogCollector(logConfig *l8logf.L8LogConfig, vnic ifs.IVNic) *LogCollector {
	return &LogCollector{logConfig: logConfig, vnic: vnic}
}

func (this LogCollector) Collect() {
	name := strings.New("agent-", this.logConfig.Path).String()
	if this.logConfig.Name == "*" {
		files, err := os.ReadDir(this.logConfig.Path)
		if err != nil {
			SendLogs(name, this.vnic, err.Error())
			return
		}
		for _, file := range files {
			subLog := &l8logf.L8LogConfig{}
			subLog.Path = this.logConfig.Path
			subLog.Name = file.Name()
			subCollector := NewLogCollector(subLog, this.vnic)
			go subCollector.Collect()
		}
		common2.WaitForSignal(this.vnic.Resources())
		return
	}

	fullPath := strings.New(this.logConfig.Path, "/", this.logConfig.Name).String()
	_, err := os.Stat(fullPath)
	for err != nil {
		fmt.Println("File '", fullPath, " does not exist ")
		time.Sleep(time.Second)
		_, err = os.Stat(fullPath)
	}
	time.Sleep(time.Second)
	err = TailFile(fullPath, this.vnic, 0)
	if err != nil {
		SendLogs(name, this.vnic, err.Error())
	}
}

func SendLogs(filename string, nic ifs.IVNic, logs ...string) {
	logF := &l8logf.L8LogF{}
	logF.SourceIp = os.Getenv("NODE_IP")
	if logF.SourceIp == "" {
		logF.SourceIp = protocol.MachineIP
	}
	logF.Filename = filename
	logF.Logs = logs
	nic.Unicast(nic.Resources().SysConfig().RemoteUuid, common.LogServiceName, common.LogServiceArea, ifs.POST, logF)
}
