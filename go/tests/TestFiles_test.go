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

// Package tests provides unit and integration tests for the L8LogFusion system.
// It includes tests for file operations, log collection, log service, and web UI.
package tests

import (
	"os"
	"testing"

	"github.com/saichler/l8logfusion/go/agent/common"
	"google.golang.org/protobuf/encoding/protojson"
)

// TestFiles verifies the FileOf function correctly scans directories and builds
// the file tree structure. It outputs the result as JSON for inspection.
func TestFiles(t *testing.T) {
	l8file := common.FileOf("/data/logdb")
	jsn, _ := protojson.Marshal(l8file)
	os.WriteFile("./example.json", jsn, 0755)
}
