package common

import (
	"os"
	"path/filepath"

	"github.com/saichler/l8logfusion/go/types/l8logf"
)

func FileOf(path string) *l8logf.L8File {
	f := &l8logf.L8File{}
	f.IsDirectory = true
	f.Files = make([]*l8logf.L8File, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	for _, file := range files {
		if file.IsDir() {
			subDir := FileOf(filepath.Join(path, file.Name()))
			if subDir == nil {
				continue
			}
			subDir.Name = file.Name()
			subDir.Path = path
			f.Files = append(f.Files, subDir)
		} else {
			sub := &l8logf.L8File{}
			sub.Name = file.Name()
			sub.Path = path
			f.Files = append(f.Files, sub)
		}
	}
	return f
}
