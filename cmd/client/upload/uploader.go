package upload

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

func (c *Command) uploadConsumer(chFiles <-chan fileInfoType, chError chan<- error) {
	// {filePath:cloth.jpeg dir:./data/}
	for file := range chFiles {
		err := c.uploadFileWithReport(file)
		if err != nil {
			chError <- err
			continue
		}
	}
	chError <- nil
}

func (c *Command) uploadFileWithReport(file fileInfoType) error {
	startT := time.Now()
	err, isDir, size, msg := c.uploadFile(file)
	cost := time.Now().UnixNano()/1000/1000 - startT.UnixNano()/1000/1000

	if err != nil {
		elog.Errorf("upload file error,file:%s,cost:%d(ms),error info:%s\n", file.filePath, cost, err.Error())
	} else {
		if file.dir == "" {
			// fix panic
			file.dir = "."
		}
		absPath := file.dir + string(os.PathSeparator) + file.filePath
		fileInfo, errF := os.Stat(absPath)
		speed := 0.0
		if cost > 0 && errF == nil {
			speed = (float64(fileInfo.Size()) / 1024) / (float64(cost) / 1000)
		}
		if errF == nil {
			elog.Infof("upload file success,file:%s,size:%d,speed:%.2f(KB/s),cost:%d(ms)\n", file.filePath, fileInfo.Size(), speed, cost)
		}
	}
	c.updateMonitor(err, isDir, size)

	c.report(msg, err)
	return err
}

func (c *Command) updateMonitor(err error, isDir bool, size int64) {
	if err != nil {
		c.monitor.updateErr(0, 1)
	} else if isDir {
		c.monitor.updateDir(size, 1)
	} else {
		c.monitor.updateFile(size, 1)
	}
	freshProgress()
}

func (c *Command) report(msg string, err error) {
	if err != nil {
		elog.Error(msg, elog.FieldErr(err))
	}
}

func (c *Command) uploadFile(file fileInfoType) (rerr error, isDir bool, size int64, msg string) {
	//first make object name
	objectName := c.makeObjectName(file)

	filePath := file.filePath
	filePath = filepath.Join(file.dir, filePath)

	rerr = nil
	isDir = false
	size = 0 // the size update to monitor

	//get file size and last modify time
	f, err := os.Stat(filePath)
	if err != nil {
		rerr = err
		return
	}

	if !f.IsDir() {
		size = f.Size()
	}

	absPath, _ := filepath.Abs(filePath)

	info, rerr := c.resty.R().SetHeaders(map[string]string{
		"clientId":     c.option.clientId,
		"clientSecret": c.option.clientSecret,
	}).SetFormData(map[string]string{
		"name":    objectName,
		"modTime": cast.ToString(f.ModTime().Unix()),
		"size":    cast.ToString(size),
	}).SetFile("myfile", absPath).Post("/upload")
	elog.Info("upload file", zap.Any("info", info))
	return
}

func (c *Command) makeObjectName(file fileInfoType) string {
	// replace "\" of file.filePath to "/"
	filePath := file.filePath
	filePath = strings.Replace(file.filePath, string(os.PathSeparator), "/", -1)
	filePath = strings.Replace(file.filePath, "\\", "/", -1)
	return filePath
}
