package logserver

import (
	"os"
	"sync"

	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8reflect/go/reflect/introspecting"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8utils/go/utils/web"
)

type LogService struct {
	files      map[string]*os.File
	mtx        *sync.Mutex
	dbLocation string
}

func NewLogService() *LogService {
	return &LogService{}
}

func (this *LogService) Activate(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) error {
	vnic.Resources().Registry().Register(&l8logf.L8LogF{})
	this.files = make(map[string]*os.File)
	this.dbLocation = sla.Args()[0].(string)
	this.mtx = &sync.Mutex{}
	err := os.MkdirAll(this.dbLocation, 0755)
	if err != nil {
		panic(err)
	}
	node, _ := vnic.Resources().Introspector().Inspect(l8logf.L8File{})
	introspecting.AddPrimaryKeyDecorator(node, "Path", "Name")
	return err
}

func (this *LogService) DeActivate() error {
	return nil
}

func (this *LogService) Post(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	for _, elem := range elements.Elements() {
		l := elem.(*l8logf.L8LogF)
		f, e := this.fetchFile(l)
		if e != nil {
			vnic.Resources().Logger().Error(e.Error())
			continue
		}
		bts := make([]byte, 0)
		for _, msg := range l.Logs {
			bts = append(bts, msg...)
		}
		n, e := f.Write(bts)
		if e != nil {
			vnic.Resources().Logger().Error(e.Error())
			continue
		}
		if n != len(bts) {
			vnic.Resources().Logger().Error("Written bytes size mismatch ", n, "!=", len(bts))
		}
	}
	return nil
}

func (this *LogService) Put(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

func (this *LogService) Patch(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

func (this *LogService) Delete(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

func (this *LogService) Get(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	q, err := elements.Query(vnic.Resources())
	if err != nil {
		return object.NewError(err.Error())
	}
	if q == nil {
		l8file := common.FileOf("/data/logdb")
		return object.New(nil, l8file)
	}
	resp, err := LoadData(q)
	return object.New(err, resp)

}

func (this *LogService) Failed(elements ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}

func (this *LogService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}

func (this *LogService) WebService() ifs.IWebService {
	ws := web.New(common.LogServiceName, common.LogServiceArea, nil,
		nil, nil, nil, nil, nil, nil, nil,
		&l8api.L8Query{}, &l8logf.L8File{})
	return ws
}
