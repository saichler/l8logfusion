package logserver

import (
	"os"
	strings2 "strings"

	"github.com/saichler/l8bus/go/overlay/vnet"
	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/strings"
)

func Main() ifs.IResources {
	r := common.NewResources("logs")
	net := vnet.NewVNet(r, true)
	net.Start()
	vnic := net.VnetVnic()
	sla := ifs.NewServiceLevelAgreement(&LogService{}, common.LogServiceName, common.LogServiceArea, true, nil)
	sla.SetArgs("/data/logdb")
	net.Resources().Services().Activate(sla, vnic)
	return r
}

func (this *LogService) fetchFile(lf *l8logf.L8LogF) (*os.File, error) {
	filename := strings.New(this.dbLocation, "/", lf.SourceIp, "/", lf.Filename).String()
	this.mtx.Lock()
	defer this.mtx.Unlock()
	f, ok := this.files[filename]
	if ok {
		return f, nil
	}
	index := strings2.LastIndex(filename, "/")
	path := filename[0:index]
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}
	f, err = os.Create(filename)
	if err != nil {
		return nil, err
	}
	this.files[filename] = f
	return f, nil
}

func (this *LogService) Files() map[string]*os.File {
	return this.files
}

func LogsService(r ifs.IResources) *LogService {
	sh, ok := r.Services().ServiceHandler(common.LogServiceName, common.LogServiceArea)
	if !ok {
		return nil
	}
	return sh.(*LogService)
}
