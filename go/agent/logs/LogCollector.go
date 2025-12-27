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

// Package logs provides the log collection agent functionality for L8LogFusion.
// It implements real-time log file monitoring, tailing, and transmission to
// the central log server over the Layer 8 virtual network overlay.
//
// The package supports:
//   - Single file monitoring with specific filename
//   - Wildcard monitoring of all .log and .err files in a directory tree
//   - Automatic file rotation detection and handling
//   - Intelligent buffering and batch transmission
//   - Crash recovery with resume from last read position
package logs

import (
	"fmt"
	common2 "github.com/saichler/probler/go/prob/common"
	"os"
	"path/filepath"
	strings2 "strings"
	"time"

	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/ipsegment"
	"github.com/saichler/l8utils/go/utils/strings"
)

// LogCollector is the main agent component responsible for monitoring and collecting
// log files from a configured path. It tails log files in real-time and transmits
// new log entries to the central log server via the Layer 8 virtual network interface.
type LogCollector struct {
	// logConfig specifies the path and filename pattern to monitor
	logConfig *l8logf.L8LogConfig
	// vnic is the virtual network interface for communicating with the log server
	vnic ifs.IVNic
}

// NewLogCollector creates a new LogCollector instance with the specified configuration.
//
// Parameters:
//   - logConfig: Configuration specifying the log path and filename (or "*" for wildcard)
//   - vnic: Virtual network interface for transmitting logs to the server
//
// Returns:
//   - *LogCollector: A configured collector ready to start monitoring
func NewLogCollector(logConfig *l8logf.L8LogConfig, vnic ifs.IVNic) *LogCollector {
	return &LogCollector{logConfig: logConfig, vnic: vnic}
}

// collect recursively scans a directory tree for log files and starts individual
// collectors for each .log or .err file found. This method is used when wildcard
// collection is configured (logConfig.Name == "*").
func (this LogCollector) collect(path, name string) {
	files, err := os.ReadDir(path)
	if err != nil {
		SendLogs(name, this.vnic, err.Error())
		return
	}
	for _, file := range files {
		if file.IsDir() {
			this.collect(filepath.Join(path, file.Name()), name)
		} else if strings2.HasSuffix(file.Name(), ".log") ||
			strings2.HasSuffix(file.Name(), ".err") {
			subLog := &l8logf.L8LogConfig{}
			subLog.Path = path
			subLog.Name = file.Name()
			subCollector := NewLogCollector(subLog, this.vnic)
			go subCollector.Collect()
		}
	}
}

// Collect starts the log collection process based on the configured path and filename.
// If the filename is "*", it recursively collects all .log and .err files in the directory.
// Otherwise, it monitors the specific file, waiting for it to exist if necessary.
//
// The method blocks until interrupted by a system signal when using wildcard collection,
// or until an error occurs when monitoring a specific file.
func (this LogCollector) Collect() {
	name := strings.New("agent-", this.logConfig.Path).String()
	if this.logConfig.Name == "*" {
		this.collect(this.logConfig.Path, name)
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

// SendLogs transmits a batch of log entries to the central log server via unicast.
// It packages the logs with source identification (IP address and filename) and
// sends them to the configured remote log service.
//
// Parameters:
//   - filename: The name of the source log file
//   - nic: The virtual network interface for transmission
//   - logs: Variable number of log line strings to send
//
// The source IP is determined from the NODE_IP environment variable, falling back
// to the machine's detected IP address if not set.
func SendLogs(filename string, nic ifs.IVNic, logs ...string) {
	logF := &l8logf.L8LogF{}
	logF.SourceIp = os.Getenv("NODE_IP")
	if logF.SourceIp == "" {
		logF.SourceIp = ipsegment.MachineIP
	}
	logF.Filename = filename
	logF.Logs = logs
	nic.Unicast(nic.Resources().SysConfig().RemoteUuid, common.LogServiceName, common.LogServiceArea, ifs.POST, logF)
}
