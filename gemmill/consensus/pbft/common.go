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

package pbft

import (
	"github.com/dappledger/AnnChain/gemmill/types"
)

// XXX: WARNING: these functions can halt the consensus as firing events is synchronous.
// Make sure to read off the channels, and in the case of subscribeToEventRespond, to write back on it

// NOTE: if chanCap=0, this blocks on the event being consumed
func subscribeToEvent(evsw types.EventSwitch, receiver, eventID string, chanCap int) chan interface{} {
	// listen for event
	ch := make(chan interface{}, chanCap)
	types.AddListenerForEvent(evsw, receiver, eventID, func(data types.TMEventData) {
		ch <- data
	})
	return ch
}

// NOTE: this blocks on receiving a response after the event is consumed
func subscribeToEventRespond(evsw types.EventSwitch, receiver, eventID string) chan interface{} {
	// listen for event
	ch := make(chan interface{})
	types.AddListenerForEvent(evsw, receiver, eventID, func(data types.TMEventData) {
		ch <- data
		<-ch
	})
	return ch
}
