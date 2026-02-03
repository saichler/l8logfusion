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

// Package main implements the L8LogFusion log collector agent Docker container entry point.
// This agent monitors log files on a node and transmits them to the central log server
// over the Layer 8 virtual network overlay.
//
// Required environment variables:
//   - NODE_IP: IP address of the log server to connect to
//   - LOGPATH: Directory path to monitor for log files
//   - LOGFILE: Specific filename to monitor, or "*" for all .log/.err files
//
// Example usage:
//
//	docker run -e NODE_IP=192.168.1.100 -e LOGPATH=/var/log -e LOGFILE="*" \
//	  -v /var/log:/var/log:ro saichler/probler-logagent:latest
package main

import (
	"os"

	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/agent/logs"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8utils/go/utils/shared"
)

// main initializes and starts the log collector agent.
// It reads configuration from environment variables, establishes a connection
// to the log server via the virtual network, and begins collecting logs.
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

	r := shared.ResourcesOf("logs", 26000, 30, true)
	r.SysConfig().RemoteVnet = ip

	nic := vnic.NewVirtualNetworkInterface(r, nil)
	nic.Start()
	nic.WaitForConnection()

	lc := &l8logf.L8LogConfig{Path: logpath, Name: logfile}
	collector := logs.NewLogCollector(lc, nic)
	collector.Collect()
}
