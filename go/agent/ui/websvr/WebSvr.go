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

// Package websvr provides the web server component for L8LogFusion's browser-based UI.
// It initializes and configures the REST server for log browsing, health monitoring,
// and query operations via the Layer 8 virtual network.
//
// Key features:
//   - REST API endpoints for log queries
//   - Health monitoring dashboard
//   - File tree navigation for browsing logs by source
//   - Integration with Layer 8 service framework
package websvr

import (
	"os"

	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8bus/go/overlay/vnic"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8health"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/ipsegment"
	"github.com/saichler/l8utils/go/utils/shared"
	"github.com/saichler/l8web/go/web/server"
	common2 "github.com/saichler/probler/go/prob/common"
)

// StartWebServer initializes and starts the REST web server for log browsing.
// It configures the server with TLS, connects to the log server via the virtual
// network, and registers all necessary web service endpoints.
//
// Parameters:
//   - port: HTTP/HTTPS port to listen on (typically 26000)
//   - cert: Path to TLS certificate directory
//
// The server:
//   - Connects to the log server on port 12443
//   - Registers health monitoring endpoints
//   - Enables web service endpoint discovery
//   - Blocks until server shutdown
func StartWebServer(port int, cert string) {
	serverConfig := &server.RestServerConfig{
		Host:           ipsegment.MachineIP,
		Port:           port,
		Authentication: false,
		CertName:       cert,
		Prefix:         common2.PREFIX,
	}
	svr, err := server.NewRestServer(serverConfig)
	if err != nil {
		panic(err)
	}

	resources := shared.ResourcesOf("logs-web-"+os.Getenv("HOSTNAME"), 12443, 30, false)
	nic := vnic.NewVirtualNetworkInterface(resources, nil)
	nic.Resources().SysConfig().KeepAliveIntervalSeconds = 60
	nic.Start()
	nic.WaitForConnection()

	nic.Resources().Registry().Register(&l8api.L8Query{})
	nic.Resources().Registry().Register(&l8health.L8Top{})
	nic.Resources().Registry().Register(&l8web.L8Empty{})
	nic.Resources().Registry().Register(&l8health.L8Health{})
	nic.Resources().Registry().Register(&l8health.L8HealthList{})
	nic.Resources().Registry().Register(&l8logf.L8File{})
	nic.Resources().Introspector().Decorators().AddPrimaryKeyDecorator(&l8logf.L8File{}, "Path", "Name")

	hs, ok := nic.Resources().Services().ServiceHandler(health.ServiceName, 0)
	if ok {
		ws := hs.WebService()
		svr.RegisterWebService(ws, nic)
	}

	//Activate the webpoints service
	sla := ifs.NewServiceLevelAgreement(&server.WebService{}, ifs.WebService, 0, false, nil)
	sla.SetArgs(svr)
	nic.Resources().Services().Activate(sla, nic)

	nic.Resources().Logger().Info("Web Server Started!")

	svr.Start()
}
