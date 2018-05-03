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

package upnp

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
)

type UPNPCapabilities struct {
	PortMapping bool
	Hairpin     bool
}

func makeUPNPListener(logger *zap.Logger, intPort int, extPort int) (NAT, net.Listener, net.IP, error) {
	nat, err := Discover()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("NAT upnp could not be discovered: %v", err)
	}
	logger.Debug(Fmt("ourIP: %v", nat.(*upnpNAT).ourIP))

	ext, err := nat.GetExternalAddress()
	if err != nil {
		return nat, nil, nil, fmt.Errorf("External address error: %v", err)
	}
	logger.Debug(Fmt("External address: %v", ext))

	port, err := nat.AddPortMapping("tcp", extPort, intPort, "Tendermint UPnP Probe", 0)
	if err != nil {
		return nat, nil, ext, fmt.Errorf("Port mapping error: %v", err)
	}
	logger.Debug(Fmt("Port mapping mapped: %v", port))

	// also run the listener, open for all remote addresses.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", intPort))
	if err != nil {
		return nat, nil, ext, fmt.Errorf("Error establishing listener: %v", err)
	}
	return nat, listener, ext, nil
}

func testHairpin(logger *zap.Logger, listener net.Listener, extAddr string) (supportsHairpin bool) {
	// Listener
	go func() {
		inConn, err := listener.Accept()
		if err != nil {
			logger.Info(Fmt("Listener.Accept() error: %v", err))
			return
		}
		logger.Debug(Fmt("Accepted incoming connection: %v -> %v", inConn.LocalAddr(), inConn.RemoteAddr()))
		buf := make([]byte, 1024)
		n, err := inConn.Read(buf)
		if err != nil {
			logger.Info(Fmt("Incoming connection read error: %v", err))
			return
		}
		logger.Debug(Fmt("Incoming connection read %v bytes: %X", n, buf))
		if string(buf) == "test data" {
			supportsHairpin = true
			return
		}
	}()

	// Establish outgoing
	outConn, err := net.Dial("tcp", extAddr)
	if err != nil {
		logger.Info(Fmt("Outgoing connection dial error: %v", err))
		return
	}

	n, err := outConn.Write([]byte("test data"))
	if err != nil {
		logger.Info(Fmt("Outgoing connection write error: %v", err))
		return
	}
	logger.Debug(Fmt("Outgoing connection wrote %v bytes", n))

	// Wait for data receipt
	time.Sleep(1 * time.Second)
	return
}

func Probe(logger *zap.Logger) (caps UPNPCapabilities, err error) {
	logger.Debug("Probing for UPnP!")

	intPort, extPort := 8001, 8001

	nat, listener, ext, err := makeUPNPListener(logger, intPort, extPort)
	if err != nil {
		return
	}
	caps.PortMapping = true

	// Deferred cleanup
	defer func() {
		err = nat.DeletePortMapping("tcp", intPort, extPort)
		if err != nil {
			logger.Warn(Fmt("Port mapping delete error: %v", err))
		}
		listener.Close()
	}()

	supportsHairpin := testHairpin(logger, listener, fmt.Sprintf("%v:%v", ext, extPort))
	if supportsHairpin {
		caps.Hairpin = true
	}

	return
}
