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

package rpctypes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/dappledger/AnnChain/gemmill/modules/go-events"
)

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func NewRPCRequest(id string, method string, params []interface{}) RPCRequest {
	return RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

//----------------------------------------

type Result interface {
}

//----------------------------------------

type RPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      string           `json:"id"`
	Result  *json.RawMessage `json:"result"`
	Error   string           `json:"error"`
}

type Options struct {
	JSONName      string      // (JSON) Corresponding JSON field name. (override with `json=""`)
	JSONOmitEmpty bool        // (JSON) Omit field if value is empty
	Varint        bool        // (Binary) Use length-prefixed encoding for (u)int64
	Unsafe        bool        // (JSON/Binary) Explicitly enable support for floats or maps
	ZeroValue     interface{} // Prototype zero object
}

var typeInfosMtx sync.RWMutex
var typeInfos = map[reflect.Type]*TypeInfo{}

type TypeInfo struct {
	Type reflect.Type // The type

	// If Type is kind reflect.Interface, is registered
	IsRegisteredInterface bool
	ByteToType            map[byte]reflect.Type
	TypeToByte            map[reflect.Type]byte

	// If Type is kind reflect.Struct
	Fields []StructFieldInfo
	Unwrap bool // if struct has only one field and its an anonymous interface
}
type StructFieldInfo struct {
	Index   int          // Struct field index
	Type    reflect.Type // Struct field type
	Options              // Encoding options
}

func (info StructFieldInfo) Unpack() (int, reflect.Type, Options) {
	return info.Index, info.Type, info.Options
}

var (
	TimeType = GetTypeFromStructDeclaration(struct{ time.Time }{})
)

const (
	Iso8601 = "2006-01-02T15:04:05.000Z" // forced microseconds
)

func GetTypeFromStructDeclaration(o interface{}) reflect.Type {
	rt := reflect.TypeOf(o)
	return rt.Field(0).Type
}
func JsonBytes(res interface{}) (bytez []byte, err error) {

	w, n := new(bytes.Buffer), new(int)
	rv := reflect.ValueOf(res)
	rt := reflect.TypeOf(res)
	if rv.Kind() == reflect.Ptr {
		rv, rt = rv.Elem(), rt.Elem()
	}
	err = writeReflectJSON(rv, rt, Options{}, w, n)
	if err != nil {
		return
	}
	bytez = w.Bytes()
	return

}
func MakeTypeInfo(rt reflect.Type) *TypeInfo {
	info := &TypeInfo{Type: rt}

	// If struct, register field name options
	if rt.Kind() == reflect.Struct {
		numFields := rt.NumField()
		structFields := []StructFieldInfo{}
		for i := 0; i < numFields; i++ {
			field := rt.Field(i)
			if field.PkgPath != "" {
				continue
			}
			skip, opts := getOptionsFromField(field)
			if skip {
				continue
			}
			structFields = append(structFields, StructFieldInfo{
				Index:   i,
				Type:    field.Type,
				Options: opts,
			})
		}
		info.Fields = structFields

		// Maybe type is a wrapper.
		if len(structFields) == 1 {
			jsonName := rt.Field(structFields[0].Index).Tag.Get("json")
			if jsonName == "unwrap" {
				info.Unwrap = true
			}
		}
	}

	return info
}
func getOptionsFromField(field reflect.StructField) (skip bool, opts Options) {
	jsonTag := field.Tag.Get("json")
	//binTag := field.Tag.Get("binary")
	//	wireTag := field.Tag.Get("wire")
	if jsonTag == "-" {
		skip = true
		return
	}
	jsonTagParts := strings.Split(jsonTag, ",")
	if jsonTagParts[0] == "" {
		opts.JSONName = field.Name
	} else {
		opts.JSONName = jsonTagParts[0]
	}
	if len(jsonTagParts) > 1 {
		if jsonTagParts[1] == "omitempty" {
			opts.JSONOmitEmpty = true
		}
	}
	/*
		if binTag == "varint" { // TODO: extend
			opts.Varint = true
		}
		if wireTag == "unsafe" {
			opts.Unsafe = true
		}
	*/
	opts.ZeroValue = reflect.Zero(field.Type).Interface()
	return
}
func GetTypeInfo(rt reflect.Type) *TypeInfo {
	typeInfosMtx.RLock()
	info := typeInfos[rt]
	typeInfosMtx.RUnlock()
	if info == nil {
		info = MakeTypeInfo(rt)
		typeInfosMtx.Lock()
		typeInfos[rt] = info
		typeInfosMtx.Unlock()
	}
	return info
}
func WriteTo(bz []byte, w io.Writer, n *int) (err error) {
	n_, err_ := w.Write(bz)
	*n += n_
	err = err_
	return
}

func isEmpty(rt reflect.Type, rv reflect.Value, opts Options) bool {
	if rt.Comparable() {
		// if its comparable we can check directly
		if rv.Interface() == opts.ZeroValue {
			return true
		}
		return false
	} else {
		// TODO: A faster alternative might be to call writeReflectJSON
		// onto a buffer and check if its "{}" or not.
		switch rt.Kind() {
		case reflect.Struct:
			// check fields
			typeInfo := GetTypeInfo(rt)
			for _, fieldInfo := range typeInfo.Fields {
				fieldIdx, fieldType, opts := fieldInfo.Unpack()
				fieldRv := rv.Field(fieldIdx)
				if !isEmpty(fieldType, fieldRv, opts) { // Skip zero value if omitempty
					return false
				}
			}
			return true

		default:
			if rv.Len() == 0 {
				return true
			}
			return false
		}
	}
	return false
}

func writeReflectJSON(rv reflect.Value, rt reflect.Type, opts Options, w io.Writer, n *int) (err error) {

	// Get typeInfo
	typeInfo := GetTypeInfo(rt)
	if rt.Kind() == reflect.Interface {
		if rv.IsNil() {
			err = WriteTo([]byte("null"), w, n)
			return
		}
		crv := rv.Elem()  // concrete reflection value
		crt := crv.Type() // concrete reflection type
		if typeInfo.IsRegisteredInterface {
			// See if the crt is registered.
			// If so, we're more restrictive.
			typeByte, ok := typeInfo.TypeToByte[crt]
			if !ok {
				switch crt.Kind() {
				case reflect.Ptr:
					err = fmt.Errorf("Unexpected pointer type %v for registered interface %v. "+
						"Was it registered as a value receiver rather than as a pointer receiver?", crt, rt.Name())
				case reflect.Struct:
					err = fmt.Errorf("Unexpected struct type %v for registered interface %v. "+
						"Was it registered as a pointer receiver rather than as a value receiver?", crt, rt.Name())
				default:
					err = fmt.Errorf("Unexpected type %v for registered interface %v. "+
						"If this is intentional, please register it.", crt, rt.Name())
				}
				return
			}
			if crt.Kind() == reflect.Ptr {
				crv, crt = crv.Elem(), crt.Elem()
				if !crv.IsValid() {
					err = fmt.Errorf("Unexpected nil-pointer of type %v for registered interface %v. "+
						"For compatibility with other languages, nil-pointer interface values are forbidden.", crt, rt.Name())
					return
				}
			}
			err = WriteTo([]byte(fmt.Sprintf("[%v,", typeByte)), w, n)
			if err != nil {
				return
			}
			err = writeReflectJSON(crv, crt, opts, w, n)
			if err != nil {
				return
			}
			err = WriteTo([]byte("]"), w, n)
		} else {
			// We support writing unregistered interfaces for convenience.
			err = writeReflectJSON(crv, crt, opts, w, n)
		}
		return
	}

	if rt.Kind() == reflect.Ptr {
		// Dereference pointer
		rv, rt = rv.Elem(), rt.Elem()
		typeInfo = GetTypeInfo(rt)
		if !rv.IsValid() {
			// For better compatibility with other languages,
			// as far as tendermint/wire is concerned,
			// pointers to nil values are the same as nil.
			err = WriteTo([]byte("null"), w, n)
			return
		}
		// continue...
	}

	// All other types
	switch rt.Kind() {
	case reflect.Array:
		elemRt := rt.Elem()
		length := rt.Len()
		if elemRt.Kind() == reflect.Uint8 {
			// Special case: Bytearray
			bytearray := reflect.ValueOf(make([]byte, length))
			reflect.Copy(bytearray, rv)
			err = WriteTo([]byte(fmt.Sprintf("\"%X\"", bytearray.Interface())), w, n)
		} else {
			err = WriteTo([]byte("["), w, n)
			if err != nil {
				return
			}
			// Write elems
			for i := 0; i < length; i++ {
				elemRv := rv.Index(i)
				err = writeReflectJSON(elemRv, elemRt, opts, w, n)
				if err != nil {
					return
				}
				if i < length-1 {
					err = WriteTo([]byte(","), w, n)
					if err != nil {
						return
					}
				}
			}
			err = WriteTo([]byte("]"), w, n)
		}

	case reflect.Slice:
		elemRt := rt.Elem()
		if elemRt.Kind() == reflect.Uint8 {
			// Special case: Byteslices
			byteslice := rv.Bytes()
			err = WriteTo([]byte(fmt.Sprintf("\"%X\"", byteslice)), w, n)
		} else {
			err = WriteTo([]byte("["), w, n)
			if err != nil {
				return
			}
			// Write elems
			length := rv.Len()
			for i := 0; i < length; i++ {
				elemRv := rv.Index(i)
				err = writeReflectJSON(elemRv, elemRt, opts, w, n)
				if err != nil {
					return
				}
				if i < length-1 {
					err = WriteTo([]byte(","), w, n)
					if err != nil {
						return
					}
				}
			}
			err = WriteTo([]byte("]"), w, n)
		}

	case reflect.Struct:
		if rt == TimeType {
			// Special case: time.Time
			t := rv.Interface().(time.Time).UTC()
			str := t.Format(Iso8601)
			jsonBytes, err_ := json.Marshal(str)
			if err_ != nil {
				err = err_
				return
			}
			err = WriteTo(jsonBytes, w, n)
		} else {
			if typeInfo.Unwrap {
				fieldIdx, fieldType, opts := typeInfo.Fields[0].Unpack()
				fieldRv := rv.Field(fieldIdx)
				err = writeReflectJSON(fieldRv, fieldType, opts, w, n)
			} else {
				err = WriteTo([]byte("{"), w, n)
				if err != nil {
					return
				}
				wroteField := false
				for _, fieldInfo := range typeInfo.Fields {
					fieldIdx, fieldType, opts := fieldInfo.Unpack()
					fieldRv := rv.Field(fieldIdx)
					if opts.JSONOmitEmpty && isEmpty(fieldType, fieldRv, opts) { // Skip zero value if omitempty
						continue
					}
					if wroteField {
						err = WriteTo([]byte(","), w, n)
						if err != nil {
							return
						}
					} else {
						wroteField = true
					}
					err = WriteTo([]byte(fmt.Sprintf("\"%v\":", opts.JSONName)), w, n)
					if err != nil {
						return
					}
					err = writeReflectJSON(fieldRv, fieldType, opts, w, n)
					if err != nil {
						return
					}
				}
				err = WriteTo([]byte("}"), w, n)
			}
		}

	case reflect.String:
		fallthrough
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		fallthrough
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		fallthrough
	case reflect.Bool:
		jsonBytes, err_ := json.Marshal(rv.Interface())
		if err_ != nil {
			err = err_
			return
		}
		err = WriteTo(jsonBytes, w, n)

	case reflect.Float64, reflect.Float32:
		if !opts.Unsafe {
			err = errors.New("Wire float* support requires `wire:\"unsafe\"`")
			return
		}
		jsonBytes, err_ := json.Marshal(rv.Interface())
		if err_ != nil {
			err = err_
			return
		}
		err = WriteTo(jsonBytes, w, n)

	default:
		err = fmt.Errorf("Unknown field type %v", rt.Kind())
	}
	return
}

func NewRPCResponse(id string, res interface{}, err string) RPCResponse {

	var raw *json.RawMessage
	if res != nil {
		bytez, errJ := JsonBytes(res)
		if errJ != nil {
			err = errJ.Error()
		} else {
			rawMsg := json.RawMessage(bytez)
			raw = &rawMsg
		}
	}
	return RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  raw,
		Error:   err,
	}
}

//----------------------------------------

// *wsConnection implements this interface.
type WSRPCConnection interface {
	GetRemoteAddr() string
	GetEventSwitch() events.EventSwitch
	WriteRPCResponse(resp RPCResponse)
	TryWriteRPCResponse(resp RPCResponse) bool
}

// websocket-only RPCFuncs take this as the first parameter.
type WSRPCContext struct {
	Request RPCRequest
	WSRPCConnection
}

//----------------------------------------
// sockets
//
// Determine if its a unix or tcp socket.
// If tcp, must specify the port; `0.0.0.0` will return incorrectly as "unix" since there's no port
func SocketType(listenAddr string) string {
	socketType := "unix"
	if len(strings.Split(listenAddr, ":")) >= 2 {
		socketType = "tcp"
	}
	return socketType
}
