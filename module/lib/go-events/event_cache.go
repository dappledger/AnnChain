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

package events

const (
	eventsBufferSize = 1000
)

// An EventCache buffers events for a Fireable
// All events are cached. Filtering happens on Flush
type EventCache struct {
	evsw   Fireable
	events []eventInfo
}

// Create a new EventCache with an EventSwitch as backend
func NewEventCache(evsw Fireable) *EventCache {
	return &EventCache{
		evsw:   evsw,
		events: make([]eventInfo, eventsBufferSize),
	}
}

// a cached event
type eventInfo struct {
	event string
	data  EventData
}

// Cache an event to be fired upon finality.
func (evc *EventCache) FireEvent(event string, data EventData) {
	// append to list
	evc.events = append(evc.events, eventInfo{event, data})
}

// Fire events by running evsw.FireEvent on all cached events. Blocks.
// Clears cached events
func (evc *EventCache) Flush() {
	for _, ei := range evc.events {
		evc.evsw.FireEvent(ei.event, ei.data)
	}
	evc.events = make([]eventInfo, eventsBufferSize)
}
