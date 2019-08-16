package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FileExists(filename string) bool {
	fi, err := os.Lstat(filename)
	if fi != nil || (err != nil && !os.IsNotExist(err)) {
		return true
	}
	return false
}

func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := HomeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func EnsureDir(dir string, mode os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, mode)
		if err != nil {
			return fmt.Errorf("could not create directory %v. %v", dir, err)
		}
	}
	return nil
}

func ReadFileData(str string) ([]byte, error) {
	path, err := PathExists(str)
	if err != nil || !path {
		fstr := strings.Replace(str, "\\\r\n", "\r\n", -1)
		fstr = strings.Replace(fstr, "\\\"", "\"", -1)
		return []byte(fstr), nil
	}
	return ioutil.ReadFile(str)
}
