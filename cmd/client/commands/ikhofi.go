package commands

import (
	_ "encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func fileData(str string) ([]byte, error) {
	path, _ := pathExists(str)
	if !path {
		fstr := strings.Replace(str, "\\\r\n", "\r\n", -1)
		fstr = strings.Replace(fstr, "\\\"", "\"", -1)
		return []byte(fstr), nil
	}
	return ioutil.ReadFile(str)
}
