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

package wire

import (
	"io"
	"math"
)

// Float32

func WriteFloat32(f float32, w io.Writer, n *int, err *error) {
	WriteUint32(math.Float32bits(f), w, n, err)
}

func ReadFloat32(r io.Reader, n *int, err *error) float32 {
	x := ReadUint32(r, n, err)
	return math.Float32frombits(x)
}

// Float64

func WriteFloat64(f float64, w io.Writer, n *int, err *error) {
	WriteUint64(math.Float64bits(f), w, n, err)
}

func ReadFloat64(r io.Reader, n *int, err *error) float64 {
	x := ReadUint64(r, n, err)
	return math.Float64frombits(x)
}
