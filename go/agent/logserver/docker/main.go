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

// Package main implements the L8LogFusion central log server Docker container entry point.
// The log server receives log entries from distributed collector agents, persists them
// to disk organized by source IP and filename, and provides query APIs for log retrieval.
//
// Usage:
//
//	docker run -v /data/logdb:/data/logdb saichler/probler-logserver:latest 12443
//
// Command-line arguments:
//   - arg[1]: Virtual network port number (e.g., 12443)
//
// The server stores logs at /data/logdb/{sourceIP}/{filename} and provides
// a REST API for querying and retrieving stored logs.
package main

import (
	"os"
	"strconv"

	"github.com/saichler/l8bus/go/overlay/vnet"
	"github.com/saichler/l8logfusion/go/agent/logserver"
	"github.com/saichler/l8utils/go/utils/logger"
	"github.com/saichler/l8utils/go/utils/shared"
)

// main initializes and starts the central log server.
// It creates a virtual network, activates the log service, and waits for
// system signals to gracefully shut down.
func main() {
	vnetPort, _ := strconv.Atoi(os.Args[1])
	r := shared.ResourcesOf("logsVnet", uint32(vnetPort), 0, false)
	vnt := vnet.NewVNet(r)
	vnt.Start()
	logserver.ActivateLogService("/data/logdb", vnt.VnetVnic())
	logger.WaitForSignal(r)
}
