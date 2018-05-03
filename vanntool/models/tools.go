/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package models

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/vanntool/tools"
)

func ParseStringArg(arg string) string {
	arg = strings.Replace(arg, "\r\n", "\\\r\n", -1)
	arg = strings.Replace(arg, "\"", "\\\"", -1)
	return arg
}

// ParseArgs parses struct_form's `form`tagname to command arguments
func ParseArgs(form interface{}, commands []string) []string {
	vl := reflect.ValueOf(form).Elem()
	args := parseArgs(reflect.TypeOf(form).Elem(), &vl)
	return append(append(commands), args...)
}

func parseArgs(tp reflect.Type, vl *reflect.Value) []string {
	argsLen := tp.NumField()
	args := make([]string, 0, argsLen*2)
	for i := 0; i < argsLen; i++ {
		if argName := tp.Field(i).Tag.Get("form"); len(argName) > 0 {
			fieldi := vl.Field(i)
			args = append(args, fmt.Sprintf("--%v", argName))
			var argValue string
			switch fieldi.Kind() {
			case reflect.Uint:
				argValue = strconv.Itoa(int(fieldi.Uint()))
			case reflect.Bool:
				argValue = strconv.FormatBool(fieldi.Bool())
			case reflect.String:
				argValue = fieldi.String()
			default:
				beego.Error("[parse_args],unparsed param:", argName)
			}
			args = append(args, argValue)
		}
	}
	return args
}

func RunShell(args []string) string {
	cmd := exec.Command(config.GetAnntoolPath(), args...)
	bytes, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%v:%v", err.Error(), string(bytes))
	}
	return string(bytes)
}

// ret: filename, filedata, error
func LoadFile(c *beego.Controller, inputID string) (string, []byte, error) {
	f, h, err := c.GetFile(inputID)
	if err != nil {
		return "", nil, err
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	if err != nil {
		return "", nil, err
	}
	return h.Filename, buf.Bytes(), err
}

// ret: filename, file temp path, error
func LoadAndSaveTempFile(c *beego.Controller, inputID string) (string, string, error) {
	fileName, fileData, err := LoadFile(c, inputID)
	if err != nil {
		return "", "", err
	}
	hashPath, err := tools.HashKeccak(fileData)
	if err != nil {
		return "", "", err
	}
	savePath := fmt.Sprintf("%v%x", config.GetJvmFilePath(), hashPath)
	err = ioutil.WriteFile(savePath, fileData, 0644)
	if err != nil {
		return "", "", err
	}
	return fileName, savePath, nil
}

func EnsureDir(path string) error {
	pathIndex := strings.LastIndex(path, "/")
	if pathIndex > 0 {
		if err := cmn.EnsureDir(path[:pathIndex], 0700); err != nil {
			return err
		}
	}
	return nil
}
