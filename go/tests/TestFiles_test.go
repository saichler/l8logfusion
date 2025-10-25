package tests

import (
	"os"
	"testing"

	"github.com/saichler/l8logfusion/go/agent/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestFiles(t *testing.T) {
	l8file := common.FileOf("/data/logdb")
	jsn, _ := protojson.Marshal(l8file)
	os.WriteFile("./example.json", jsn, 0755)
}
