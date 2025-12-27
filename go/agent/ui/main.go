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

// Package main implements the L8LogFusion web UI Docker container entry point.
// The web UI provides a browser-based interface for browsing and querying
// collected logs from the central log server.
//
// The web server runs on port 26000 by default and connects to the log server
// via the Layer 8 virtual network overlay on port 12443.
//
// Usage:
//
//	docker run -p 26000:26000 -v /data/probler:/data/probler saichler/probler-logsui:latest
package main

import (
	"github.com/saichler/l8logfusion/go/agent/ui/websvr"
)

// main starts the web server for log browsing and monitoring.
// It initializes the REST server on port 26000 with TLS certificates
// from /data/probler.
func main() {
	websvr.StartWebServer(26000, "/data/probler")
}
