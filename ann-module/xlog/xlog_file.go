// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File I/O for logs.

package xlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MaxSize is the maximum size of a log file in bytes.
var MaxSize uint64 = 1024 * 1024 * 1800

// logDirs lists the candidate directories for new log files.
var logDirs []string

// If non-empty, overrides the choice of directory in which to write logs.
// See createLogDirs for the full list of possible destinations.
//var logDir = flag.String("log_dir", "", "./")
var logDir = "./"

func createLogDirs() {
	if logDir != "" {
		logDirs = append(logDirs, logDir)
	} else {
		logDirs = append(logDirs, "./log")
	}
}

var (
	pid     = os.Getpid()
	program = filepath.Base(os.Args[0])

//	host     = "unknownhost"
//	userName = "unknownuser"
)

func logTagName(tag string) string {
	switch tag {
	case "DEBUG":
		return "dbg"
	case "INFO":
		return "log"
	case "WORNING":
		return "log"
	case "ERROR":
		return "err"
	case "FATAL":
		return "err"
	}
	return "dbg"
}

func dateString(t time.Time) string {
	return fmt.Sprintf("%04d%02d%02d-%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
	)
}

// logName returns a new log file name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(tag string, t time.Time) (name, link string) {
	//	name = fmt.Sprintf("%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
	//		program,
	//		host,
	//		userName,
	//		tag,
	//		t.Year(),
	//		t.Month(),
	//		t.Day(),
	//		t.Hour(),
	//		t.Minute(),
	//		t.Second(),
	//		pid)
	name = fmt.Sprintf("%s.%s.%s",
		program,
		tag,
		dateString(t),
	)
	return name, program + "." + tag
}

var onceLogDirs sync.Once

// create creates a new log file and returns the file and its filename, which
// contains tag ("INFO", "FATAL", etc.) and t.  If the file is created
// successfully, create also attempts to update the symlink for that tag, ignoring
// errors.
func create(tag string, t time.Time) (f *os.File, filename string, err error) {
	onceLogDirs.Do(createLogDirs)
	if len(logDirs) == 0 {
		return nil, "", errors.New("log: no log dirs")
	}
	name, link := logName(logTagName(tag), t)
	var lastErr error
	for _, dir := range logDirs {
		fname := filepath.Join(dir, name)
		f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			symlink := filepath.Join(dir, link)
			os.Remove(symlink)        // ignore err
			os.Symlink(name, symlink) // ignore err
			return f, fname, nil
		}
		lastErr = err
	}
	return nil, "", fmt.Errorf("log: cannot create log: %v", lastErr)
}
