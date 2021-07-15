package upload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
)

const (
	ChannelBuf int = 1000
)

type Command struct {
	monitor    Monitor
	option     option
	resty      *ehttp.Component
	srcDirList []string
}

var (
	mu               sync.RWMutex // mu is the mutex for interacting with user
	chProgressSignal chan chProgressSignalType
	signalNum        = 0
)

type chProgressSignalType struct {
	finish   bool
	exitStat int
}

func RunCommand(clientId, clientSecret, addr string, dirs []string, debug bool, readTimeout time.Duration) error {
	c := &Command{
		option: option{
			routines:     3,
			clientId:     clientId,
			clientSecret: clientSecret,
		},
	}
	c.monitor = Monitor{}
	c.monitor.init()
	c.resty = ehttp.DefaultContainer().Build(
		ehttp.WithReadTimeout(readTimeout),
		ehttp.WithDebug(debug),
		ehttp.WithAddr(addr),
	)
	chProgressSignal = make(chan chProgressSignalType, 10)
	go c.progressBar()
	//get file list
	srcDirList, err := c.getStorageDirs(dirs)
	if err != nil {
		return err
	}
	c.srcDirList = srcDirList
	return c.uploadFiles()
}

func (c *Command) getStorageDirs(dirs []string) ([]string, error) {
	dirList := make([]string, 0)
	for _, url := range dirs {
		storageURL, err := fileURLString(url)
		if err != nil {
			return dirList, err
		}
		dirList = append(dirList, storageURL)
	}
	return dirList, nil
}

func (c *Command) progressBar() {
	// fetch all reveal
	for signal := range chProgressSignal {
		elog.Info(c.monitor.progressBar(signal.finish, signal.exitStat))
	}
}

func fileURLString(urlStr string) (string, error) {
	if len(urlStr) >= 2 && urlStr[:2] == "~"+string(os.PathSeparator) {
		homeDir := currentHomeDir()
		if homeDir != "" {
			urlStr = strings.Replace(urlStr, "~", homeDir, 1)
		} else {
			return "", fmt.Errorf("current home dir is empty")
		}
	}
	return urlStr, nil
}

type option struct {
	addr           string
	debug          bool
	clientId       string
	clientSecret   string
	timeout        time.Duration
	routines       int64
	onlyCurrentDir bool // 表示仅操作当前目录下的文件或者object, 忽略子目录
	filters        []filterOptionType
}

type filterOptionType struct {
	name    string
	pattern string
}

type fileInfoType struct {
	filePath string
	dir      string
}

func (c *Command) uploadFiles() error {
	// producer list files
	// consumer set acl
	chFiles := make(chan fileInfoType, ChannelBuf)
	chError := make(chan error, c.option.routines)
	chListError := make(chan error, 1)
	go c.fileStatistic(c.srcDirList)
	go c.fileProducer(chFiles, chListError)

	elog.Infof("upload files,routin count:%d,multi part size\n", c.option.routines)
	for i := 0; int64(i) < c.option.routines; i++ {
		go c.uploadConsumer(chFiles, chError)
	}

	completed := 0
	for int64(completed) <= c.option.routines {
		select {
		case err := <-chListError:
			if err != nil {
				return err
			}
			completed++
		case err := <-chError:
			if err == nil {
				completed++
			}
		}
	}
	c.closeProgress()
	elog.Info(c.monitor.progressBar(true, normalExit))
	return nil
}

func (c *Command) fileStatistic(srcURLList []string) {
	for _, name := range srcURLList {
		f, err := os.Stat(name)
		if err != nil {
			c.monitor.setScanError(err)
			return
		}
		if f.IsDir() {
			if !strings.HasSuffix(name, string(os.PathSeparator)) {
				// for link directory
				name += string(os.PathSeparator)
			}

			err := c.getFileListStatistic(name)
			if err != nil {
				c.monitor.setScanError(err)
				return
			}
		} else {
			c.monitor.updateScanSizeNum(f.Size(), 1)
		}
	}

	c.monitor.setScanEnd()
	freshProgress()
}

func (c *Command) getFileListStatistic(dpath string) error {
	if c.option.onlyCurrentDir {
		return c.getCurrentDirFilesStatistic(dpath)
	}

	symlinkDiretorys := []string{dpath}
	walkFunc := func(fpath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		realFileSize := f.Size()
		dpath = filepath.Clean(dpath)
		fpath = filepath.Clean(fpath)
		_, err = filepath.Rel(dpath, fpath)
		if err != nil {
			return fmt.Errorf("list file error: %s, info: %s", fpath, err.Error())
		}

		if f.IsDir() {
			return nil
		}

		// link file or link dir
		if f.Mode()&os.ModeSymlink != 0 {
			// there is difference between os.Stat and os.Lstat in filepath.Walk
			realInfo, err := os.Stat(fpath)
			if err != nil {
				return err
			}

			if realInfo.IsDir() {
				realFileSize = 0
			} else {
				realFileSize = realInfo.Size()
			}
		}
		c.monitor.updateScanSizeNum(realFileSize, 1)
		return nil
	}

	var err error
	for {
		symlinks := symlinkDiretorys
		symlinkDiretorys = []string{}
		for _, v := range symlinks {
			err = filepath.Walk(v, walkFunc)
			if err != nil {
				return err
			}
		}
		if len(symlinkDiretorys) == 0 {
			break
		}
	}
	return err
}

func (c *Command) getCurrentDirFilesStatistic(dpath string) error {
	if !strings.HasSuffix(dpath, string(os.PathSeparator)) {
		dpath += string(os.PathSeparator)
	}

	fileList, err := ioutil.ReadDir(dpath)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileList {
		if !fileInfo.IsDir() {
			realInfo, errF := os.Stat(dpath + fileInfo.Name())
			if errF == nil && realInfo.IsDir() {
				// for symlink
				continue
			}

			c.monitor.updateScanSizeNum(fileInfo.Size(), 1)
		}
	}
	return nil
}

func freshProgress() {
	if len(chProgressSignal) <= signalNum {
		chProgressSignal <- chProgressSignalType{false, normalExit}
	}
}

func (c *Command) fileProducer(chFiles chan<- fileInfoType, chListError chan<- error) {
	for _, name := range c.srcDirList {
		f, err := os.Stat(name)
		if err != nil {
			chListError <- err
			return
		}
		if f.IsDir() {
			if !strings.HasSuffix(name, string(os.PathSeparator)) {
				// for link directory
				name += string(os.PathSeparator)
			}

			err := c.getFileList(name, chFiles)
			if err != nil {
				chListError <- err
				return
			}
		} else {
			dir, fname := filepath.Split(name)
			chFiles <- fileInfoType{fname, dir}
		}
	}
	chListError <- nil
	close(chFiles)
}

func (c *Command) getFileList(dpath string, chFiles chan<- fileInfoType) error {
	if c.option.onlyCurrentDir {
		return c.getCurrentDirFileList(dpath, chFiles)
	}
	name := dpath
	symlinkDiretorys := []string{dpath}
	walkFunc := func(fpath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		dpath = filepath.Clean(dpath)
		fpath = filepath.Clean(fpath)
		fileName, err := filepath.Rel(dpath, fpath)
		if err != nil {
			return fmt.Errorf("list file error: %s, info: %s", fpath, err.Error())
		}

		// 如果为目录，不在传递
		if f.IsDir() {
			return nil
		}

		chFiles <- fileInfoType{fileName, name}
		return nil
	}

	var err error
	for {
		symlinks := symlinkDiretorys
		symlinkDiretorys = []string{}
		for _, v := range symlinks {
			err = filepath.Walk(v, walkFunc)
			if err != nil {
				return err
			}
		}
		if len(symlinkDiretorys) == 0 {
			break
		}
	}
	return err
}

func (c *Command) getCurrentDirFileList(dpath string, chFiles chan<- fileInfoType) error {
	if !strings.HasSuffix(dpath, string(os.PathSeparator)) {
		dpath += string(os.PathSeparator)
	}

	fileList, err := ioutil.ReadDir(dpath)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileList {
		if !fileInfo.IsDir() {
			realInfo, errF := os.Stat(dpath + fileInfo.Name())
			if errF == nil && realInfo.IsDir() {
				// for symlink
				continue
			}

			chFiles <- fileInfoType{fileInfo.Name(), dpath}
		}
	}
	return nil
}

func (c *Command) closeProgress() {
	signalNum = -1
}
