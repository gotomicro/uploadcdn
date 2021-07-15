package upload

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

func max(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

func getSizeString(size int64) string {
	prefix := ""
	str := fmt.Sprintf("%d", size)
	if size < 0 {
		prefix = "-"
		str = str[1:]
	}
	len := len(str)
	strList := []string{}
	i := len % 3
	if i != 0 {
		strList = append(strList, str[0:i])
	}
	for ; i < len; i += 3 {
		strList = append(strList, str[i:i+3])
	}
	return fmt.Sprintf("%s%s", prefix, strings.Join(strList, ","))
}

func currentHomeDir() string {
	homeDir := ""
	homeDrive := os.Getenv("HOMEDRIVE")
	homePath := os.Getenv("HOMEPATH")
	if runtime.GOOS == "windows" && homeDrive != "" && homePath != "" {
		homeDir = homeDrive + string(os.PathSeparator) + homePath
	}

	if homeDir != "" {
		return homeDir
	}

	usr, _ := user.Current()
	if usr != nil {
		homeDir = usr.HomeDir
	} else {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
