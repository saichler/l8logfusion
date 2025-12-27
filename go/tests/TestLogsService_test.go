/*
 * © 2025 Sharon Aicler (saichler@gmail.com)
 *
 * Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/saichler/l8bus/go/overlay/vnet"
	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/agent/logs"
	"github.com/saichler/l8logfusion/go/agent/logserver"
	"github.com/saichler/l8logfusion/go/agent/ui/websvr"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils"
	"github.com/saichler/l8utils/go/utils/ipsegment"
	"github.com/saichler/l8utils/go/utils/logger"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// vnetPort is the virtual network port used for test communication.
const (
	vnetPort = uint16(12443)
)

// startVnet creates and starts a virtual network with the log service activated.
// Returns the running VNet instance for test coordination.
func startVnet() *vnet.VNet {
	r := utils.NewResources("logsVnet", vnetPort, 0)
	vnt := vnet.NewVNet(r)
	vnt.Start()
	logserver.ActivateLogService(vnt.VnetVnic())
	return vnt
}

// startNic creates a virtual network interface and starts a log collector.
// It connects to the test VNet and begins collecting logs from the specified directory.
//
// Parameters:
//   - logDir: Directory to create and monitor for logs
//   - logFile: Specific log filename to monitor
//
// Returns:
//   - ifs.IVNic: The initialized virtual network interface
func startNic(logDir, logFile string) ifs.IVNic {
	os.MkdirAll(logDir, 0755)
	r := utils.NewResources("logs", vnetPort, 0)
	r.SysConfig().RemoteVnet = ipsegment.MachineIP
	nic := vnic.NewVirtualNetworkInterface(r, nil)
	nic.Start()
	nic.WaitForConnection()
	fmt.Println("Starting test")
	lc := &l8logf.L8LogConfig{Path: logDir, Name: logFile}
	collector := logs.NewLogCollector(lc, nic)
	go collector.Collect()
	return nic
}

// TestLogsService is an integration test that validates the complete log collection pipeline.
// It creates a VNet with the log service, starts a collector, writes test logs,
// and verifies that logs are correctly transmitted and stored on the server.
//
// The test:
//  1. Starts a virtual network with the log service
//  2. Creates a collector connected to the VNet
//  3. Writes 1000 test log entries
//  4. Verifies the server received all log entries
//  5. Tests the query API for log retrieval
//  6. Starts the web server for manual verification
func TestLogsService(t *testing.T) {
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
	nic.Resources().Introspector().Decorators().AddPrimaryKeyDecorator(&l8logf.L8File{}, "Path", "Name")

	elemsS, e := object.NewQuery("select * from l8file where path=/data/logdb/192.168.86.220/logs and name = log.log limit 100 page 0", nic.Resources())
	if e != nil {
		nic.Resources().Logger().Fail(t, "query logs failed: ", e)
		return
	}

	elems := elemsS.(*object.Elements)

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
