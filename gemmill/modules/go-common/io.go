// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
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

package common

import (
	"bytes"
	"errors"
	"io"
)

type PrefixedReader struct {
	Prefix []byte
	reader io.Reader
}

func NewPrefixedReader(prefix []byte, reader io.Reader) *PrefixedReader {
	return &PrefixedReader{prefix, reader}
}

func (pr *PrefixedReader) Read(p []byte) (n int, err error) {
	if len(pr.Prefix) > 0 {
		read := copy(p, pr.Prefix)
		pr.Prefix = pr.Prefix[read:]
		return read, nil
	} else {
		return pr.reader.Read(p)
	}
}

// NOTE: Not goroutine safe
type BufferCloser struct {
	bytes.Buffer
	Closed bool
}

func NewBufferCloser(buf []byte) *BufferCloser {
	return &BufferCloser{
		*bytes.NewBuffer(buf),
		false,
	}
}

func (bc *BufferCloser) Close() error {
	if bc.Closed {
		return errors.New("BufferCloser already closed")
	}
	bc.Closed = true
	return nil
}

func (bc *BufferCloser) Write(p []byte) (n int, err error) {
	if bc.Closed {
		return 0, errors.New("Cannot write to closed BufferCloser")
	}
	return bc.Buffer.Write(p)
}

func (bc *BufferCloser) WriteByte(c byte) error {
	if bc.Closed {
		return errors.New("Cannot write to closed BufferCloser")
	}
	return bc.Buffer.WriteByte(c)
}

func (bc *BufferCloser) WriteRune(r rune) (n int, err error) {
	if bc.Closed {
		return 0, errors.New("Cannot write to closed BufferCloser")
	}
	return bc.Buffer.WriteRune(r)
}

func (bc *BufferCloser) WriteString(s string) (n int, err error) {
	if bc.Closed {
		return 0, errors.New("Cannot write to closed BufferCloser")
	}
	return bc.Buffer.WriteString(s)
}
