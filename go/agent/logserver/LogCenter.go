package logserver

import (
	"io"
	"os"
	"path/filepath"
	strings2 "strings"

	"github.com/saichler/l8logfusion/go/agent/common"
	"github.com/saichler/l8logfusion/go/types/l8logf"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/strings"
)

func ActivateLogService(vnic ifs.IVNic) {
	sla := ifs.NewServiceLevelAgreement(&LogService{}, common.LogServiceName, common.LogServiceArea, true, nil)
	sla.SetArgs("/data/logdb")
	vnic.Resources().Services().Activate(sla, vnic)
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

func LoadData(q ifs.IQuery) (*l8logf.L8File, error) {
	path := q.ValueForParameter("path")
	name := q.ValueForParameter("name")
	l8file := &l8logf.L8File{}
	l8file.Path = path
	l8file.Name = name
	l8file.Data = &l8logf.L8FileData{}

	// KB-based paging: 5KB per page
	const bytesPerPage = 5120 // 5KB
	l8file.Data.Page = q.Page()

	// Get file info to determine total size
	filePath := filepath.Join(l8file.Path, l8file.Name)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return l8file, err
	}
	l8file.Data.Size = int32(fileInfo.Size())

	// Calculate byte offset based on page number
	offset := int64(l8file.Data.Page) * bytesPerPage

	// If offset is beyond file size, return empty content
	if offset >= fileInfo.Size() {
		l8file.Data.Content = ""
		return l8file, nil
	}

	// Open file and seek to the offset
	file, err := os.Open(filePath)
	if err != nil {
		return l8file, err
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		return l8file, err
	}

	// Read up to bytesPerPage bytes
	buffer := make([]byte, bytesPerPage)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return l8file, err
	}

	// Set the actual content read
	l8file.Data.Content = string(buffer[:n])
	return l8file, nil
}
