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

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	rpctypes "github.com/dappledger/AnnChain/gemmill/rpc/types"
)

// TODO: Deprecate support for IP:PORT or /path/to/socket
func makeHTTPDialer(remoteAddr string) (string, func(string, string) (net.Conn, error)) {

	parts := strings.SplitN(remoteAddr, "://", 2)
	var protocol, address string
	if len(parts) != 2 {
		protocol = rpctypes.SocketType(remoteAddr)
		address = remoteAddr
	} else {
		protocol, address = parts[0], parts[1]
	}

	trimmedAddress := strings.Replace(address, "/", ".", -1) // replace / with . for http requests (dummy domain)
	return trimmedAddress, func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func makeHTTPClient(remoteAddr string) (string, *http.Client) {
	address, dialer := makeHTTPDialer(remoteAddr)
	return "http://" + address, &http.Client{
		Transport: &http.Transport{
			Dial: dialer,
		},
		Timeout: time.Second * 5,
	}
}

//------------------------------------------------------------------------------------

type Client interface {
}

//------------------------------------------------------------------------------------

// JSON rpc takes params as a slice
type ClientJSONRPC struct {
	address string
	client  *http.Client
}

func NewClientJSONRPC(remote string) *ClientJSONRPC {
	address, client := makeHTTPClient(remote)
	return &ClientJSONRPC{
		address: address,
		client:  client,
	}
}

func (c *ClientJSONRPC) Call(method string, params []interface{}, result interface{}) (interface{}, error) {
	return c.call(method, params, result)
}

func (c *ClientJSONRPC) call(method string, params []interface{}, result interface{}) (interface{}, error) {
	// Make request and get responseBytes
	request := rpctypes.RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      "",
	}
	requestBytes := wire.JSONBytes(request)
	requestBuf := bytes.NewBuffer(requestBytes)

	req, err := http.NewRequest("POST", c.address, requestBuf)
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Set("contentType", "text/json")
	httpResponse, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()
	responseBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		fmt.Println("ioutilReadAll err", err)
		return nil, err
	}
	return unmarshalResponseBytes(responseBytes, result)
}

//-------------------------------------------------------------

// URI takes params as a map
type ClientURI struct {
	address string
	client  *http.Client
}

func NewClientURI(remote string) *ClientURI {
	address, client := makeHTTPClient(remote)
	return &ClientURI{
		address: address,
		client:  client,
	}
}

func (c *ClientURI) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	return c.call(method, params, result)
}

func (c *ClientURI) call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	values, err := argsToURLValues(params)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostForm(c.address+"/"+method, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return unmarshalResponseBytes(responseBytes, &result)
}

//------------------------------------------------
func ReadJSONObjectPtr(o interface{}, bytes []byte) (result interface{}, err error) {

	var object interface{}
	err = json.Unmarshal(bytes, &object)
	if err != nil {
		return
	}

	rv, rt := reflect.ValueOf(o), reflect.TypeOf(o)
	if rv.Kind() == reflect.Ptr {
		err = readReflectJSON(rv.Elem(), rt.Elem(), rpctypes.Options{}, object)
	} else {
		err = errors.New("ReadJSON(Object)Ptr expects o to be a pointer")
	}
	return
}

func readByteJSON(o interface{}) (typeByte byte, rest interface{}, err error) {
	oSlice, ok := o.([]interface{})
	if !ok {
		err = fmt.Errorf("Expected type [Byte,?] but got type %v", reflect.TypeOf(o))
		return
	}
	if len(oSlice) != 2 {
		err = fmt.Errorf("Expected [Byte,?] len 2 but got len %v", len(oSlice))
		return
	}
	typeByte_, ok := oSlice[0].(float64)
	typeByte = byte(typeByte_)
	rest = oSlice[1]
	return
}

func readReflectJSON(rv reflect.Value, rt reflect.Type, opts rpctypes.Options, o interface{}) (err error) {

	// Get typeInfo
	typeInfo := rpctypes.GetTypeInfo(rt)

	if rt.Kind() == reflect.Interface {
		if !typeInfo.IsRegisteredInterface {
			// There's no way we can read such a thing.
			err = fmt.Errorf("Cannot read unregistered interface type %v", rt)
			return
		}
		if o == nil {
			return // nil
		}
		typeByte, rest, err_ := readByteJSON(o)
		if err_ != nil {
			err = err_
			return
		}
		crt, ok := typeInfo.ByteToType[typeByte]
		if !ok {
			err = fmt.Errorf("Byte %X not registered for interface %v", typeByte, rt)
			return
		}
		if crt.Kind() == reflect.Ptr {
			crt = crt.Elem()
			crv := reflect.New(crt)
			err = readReflectJSON(crv.Elem(), crt, opts, rest)
			if err != nil {
				return
			}
			rv.Set(crv) // NOTE: orig rv is ignored.
		} else {
			crv := reflect.New(crt).Elem()
			err = readReflectJSON(crv, crt, opts, rest)
			if err != nil {
				return
			}
			rv.Set(crv) // NOTE: orig rv is ignored.
		}
		return
	}

	if rt.Kind() == reflect.Ptr {
		if o == nil {
			return // nil
		}
		// Create new struct if rv is nil.
		if rv.IsNil() {
			newRv := reflect.New(rt.Elem())
			rv.Set(newRv)
			rv = newRv
		}
		// Dereference pointer
		rv, rt = rv.Elem(), rt.Elem()
		typeInfo = rpctypes.GetTypeInfo(rt)
		// continue...
	}

	switch rt.Kind() {
	case reflect.Array:
		elemRt := rt.Elem()
		length := rt.Len()
		if elemRt.Kind() == reflect.Uint8 {
			// Special case: Bytearrays
			oString, ok := o.(string)
			if !ok {
				err = fmt.Errorf("Expected string but got type %v", reflect.TypeOf(o))
				return
			}
			buf, err_ := hex.DecodeString(oString)
			if err_ != nil {
				err = err_
				return
			}
			if len(buf) != length {
				err = fmt.Errorf("Expected bytearray of length %v but got %v", length, len(buf))
				return
			}
			reflect.Copy(rv, reflect.ValueOf(buf))
		} else {
			oSlice, ok := o.([]interface{})
			if !ok {
				err = fmt.Errorf("Expected array of %v but got type %v", rt, reflect.TypeOf(o))
				return
			}
			if len(oSlice) != length {
				err = fmt.Errorf("Expected array of length %v but got %v", length, len(oSlice))
				return
			}
			for i := 0; i < length; i++ {
				elemRv := rv.Index(i)
				err = readReflectJSON(elemRv, elemRt, opts, oSlice[i])
				if err != nil {
					return
				}
			}
		}

	case reflect.Slice:
		elemRt := rt.Elem()
		if elemRt.Kind() == reflect.Uint8 {
			// Special case: Byteslices
			oString, ok := o.(string)
			if !ok {
				err = fmt.Errorf("Expected string but got type %v", reflect.TypeOf(o))
				return
			}
			byteslice, err_ := hex.DecodeString(oString)
			if err_ != nil {
				err = err_
				return
			}
			rv.Set(reflect.ValueOf(byteslice))
		} else {
			// Read length
			oSlice, ok := o.([]interface{})
			if !ok {
				err = fmt.Errorf("Expected array of %v but got type %v", rt, reflect.TypeOf(o))
				return
			}
			length := len(oSlice)
			sliceRv := reflect.MakeSlice(rt, length, length)
			// Read elems
			for i := 0; i < length; i++ {
				elemRv := sliceRv.Index(i)
				err = readReflectJSON(elemRv, elemRt, opts, oSlice[i])
				if err != nil {
					return
				}
			}
			rv.Set(sliceRv)
		}

	case reflect.Struct:
		if rt == rpctypes.TimeType {
			// Special case: time.Time
			str, ok := o.(string)
			if !ok {
				err = fmt.Errorf("Expected string but got type %v", reflect.TypeOf(o))
				return
			}
			t, err_ := time.Parse(rpctypes.Iso8601, str)
			if err_ != nil {
				err = err_
				return
			}
			rv.Set(reflect.ValueOf(t))
		} else {
			if typeInfo.Unwrap {
				fieldIdx, fieldType, opts := typeInfo.Fields[0].Unpack()
				fieldRv := rv.Field(fieldIdx)
				err = readReflectJSON(fieldRv, fieldType, opts, o)
				if err != nil {
					return
				}
			} else {
				oMap, ok := o.(map[string]interface{})
				if !ok {
					err = fmt.Errorf("Expected map but got type %v", reflect.TypeOf(o))
					return
				}
				// TODO: ensure that all fields are set?
				// TODO: disallow unknown oMap fields?
				for _, fieldInfo := range typeInfo.Fields {
					fieldIdx, fieldType, opts := fieldInfo.Unpack()
					value, ok := oMap[opts.JSONName]
					if !ok {
						continue // Skip missing fields.
					}
					fieldRv := rv.Field(fieldIdx)
					err = readReflectJSON(fieldRv, fieldType, opts, value)
					if err != nil {
						return
					}
				}
			}

		}
	case reflect.String:
		str, ok := o.(string)
		if !ok {
			err = fmt.Errorf("Expected string but got type %v", reflect.TypeOf(o))
			return
		}
		rv.SetString(str)

	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		num, ok := o.(float64)
		if !ok {
			err = fmt.Errorf("Expected numeric but got type %v", reflect.TypeOf(o))
			return
		}
		rv.SetInt(int64(num))

	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		num, ok := o.(float64)
		if !ok {
			err = fmt.Errorf("Expected numeric but got type %v", reflect.TypeOf(o))
			return
		}
		if num < 0 {
			err = fmt.Errorf("Expected unsigned numeric but got %v", num)
			return
		}
		rv.SetUint(uint64(num))

	case reflect.Float64, reflect.Float32:
		if !opts.Unsafe {
			err = errors.New("Wire float* support requires `wire:\"unsafe\"`")
			return
		}
		num, ok := o.(float64)
		if !ok {
			err = fmt.Errorf("Expected numeric but got type %v", reflect.TypeOf(o))
			return
		}
		rv.SetFloat(num)

	case reflect.Bool:
		bl, ok := o.(bool)
		if !ok {
			err = fmt.Errorf("Expected boolean but got type %v", reflect.TypeOf(o))
			return
		}
		rv.SetBool(bl)

	default:
		err = fmt.Errorf("Unknown field type %v", rt.Kind())
	}
	return
}

func unmarshalResponseBytes(responseBytes []byte, result interface{}) (interface{}, error) {

	var err error
	response := &rpctypes.RPCResponse{}
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling rpc response: %v", err)
	}
	errorStr := response.Error
	if errorStr != "" {
		return nil, fmt.Errorf("Response error: %v", errorStr)
	}
	result, err = ReadJSONObjectPtr(result, *response.Result)
	if err != nil {
		return nil, err
	}
	return response.Result, nil
}

func argsToURLValues(args map[string]interface{}) (url.Values, error) {
	values := make(url.Values)
	if len(args) == 0 {
		return values, nil
	}
	err := argsToJson(args)
	if err != nil {
		return nil, err
	}
	for key, val := range args {
		values.Set(key, val.(string))
	}
	return values, nil
}

func argsToJson(args map[string]interface{}) error {
	var n int
	var err error
	for k, v := range args {
		// Convert byte slices to "0x"-prefixed hex
		byteSlice, isByteSlice := reflect.ValueOf(v).Interface().([]byte)
		if isByteSlice {
			args[k] = fmt.Sprintf("0x%X", byteSlice)
			continue
		}

		// Pass everything else to go-wire
		buf := new(bytes.Buffer)
		wire.WriteJSON(v, buf, &n, &err)
		if err != nil {
			return err
		}
		args[k] = buf.String()
	}
	return nil
}
