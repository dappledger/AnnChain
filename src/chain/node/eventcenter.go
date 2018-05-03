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


package node

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tendermint/tmlibs/db"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

const (
	eventCenterStateKey = "eventwarehouse-state"
)

// EventWarehouse stores those event data which are published but still not fetched by their target subchains.
//
// EventWarehouse.database is the storage.
// EventWarehouse.state is kinda like index for the storage.
type EventWarehouse struct {
	database db.DB
	state    *EventWarehouseState
}

// EventWarehouseState contains subscriberID-publisherID : []heights
type EventWarehouseState struct {
	state map[string][]def.INT
}

// NewEventWarehouse constructs a new EventWarehouse and returns a pointer to the instance
func NewEventWarehouse(database db.DB) *EventWarehouse {
	return &EventWarehouse{
		database: database,
		state:    newWarehouseState(),
	}
}

func newWarehouseState() *EventWarehouseState {
	return &EventWarehouseState{
		state: make(map[string][]def.INT),
	}
}

// GenerateID generates wellformed key used in EventWarehouse
func (ew *EventWarehouse) GenerateID(subID, pubID string, height def.INT) []byte {
	return []byte(fmt.Sprintf("%s-%s-%d", subID, pubID, height))
}

// Batch just wraps about db.NewBatch
func (ew *EventWarehouse) Batch() db.Batch {
	return ew.database.NewBatch()
}

// Push will add new record to both EventWarehouseState and EventWarehouse
func (ew *EventWarehouse) Push(batch db.Batch, subID, pubID string, height def.INT, data []byte) {
	// heights := ec.state.Append(subID, pubID, height)
	batch.Set(ew.GenerateID(subID, pubID, height), data)
	ew.state.Append(subID, pubID, height)
}

func (ew *EventWarehouse) Fetch(subID, pubID string, height def.INT) ([]byte, error) {
	heights, err := ew.state.Get(subID, pubID)
	if err != nil {
		return nil, err
	}

	for i, h := range heights {
		if h == height {
			data := ew.database.Get(ew.GenerateID(subID, pubID, height))
			ew.state.Set(subID, pubID, append(heights[:i], heights[i+1:]...))
			return data, nil
		}
	}

	return nil, errors.Errorf("(sub: %s, pub: %s, height:%d) is invalid", subID, pubID, height)
}

// Pop makes sure the one event can only be fetched once by deleting after fetched
func (ew *EventWarehouse) Pop(subID, pubID string, height def.INT) ([]byte, error) {
	heights, err := ew.state.Get(subID, pubID)
	if err != nil {
		return nil, err
	}

	for i, h := range heights {
		if h == height {
			data := ew.database.Get(ew.GenerateID(subID, pubID, height))

			ew.state.Set(subID, pubID, append(heights[:i], heights[i+1:]...))
			ew.database.Delete(ew.GenerateID(subID, pubID, height))
			return data, nil
		}
	}

	return nil, fmt.Errorf("(sub: %s, pub: %s, height:%d) is invalid", subID, pubID, height)
}

// Flush persists db writes and ec.state changes
func (ew *EventWarehouse) Flush(batch db.Batch) {
	batch.Write()
	bs, _ := ew.state.ToBytes()
	ew.database.SetSync([]byte(eventCenterStateKey), bs)
}

func (ews *EventWarehouseState) key(subID, pubID string) string {
	return fmt.Sprintf("%s-%s", subID, pubID)
}

func (ews *EventWarehouseState) Get(subID, pubID string) ([]def.INT, error) {
	key := ews.key(subID, pubID)
	s, ok := ews.state[key]
	if !ok {
		return nil, fmt.Errorf("%s doesn't exists", key)
	}
	return s, nil
}

func (ews *EventWarehouseState) Set(subID, pubID string, heights []def.INT) {
	key := ews.key(subID, pubID)
	ews.state[key] = heights
}

func (ews *EventWarehouseState) Append(subID, pubID string, height def.INT) []def.INT {
	heights, err := ews.Get(subID, pubID)
	if err != nil {
		ews.Set(subID, pubID, []def.INT{height})
		return []def.INT{height}
	}
	heights = append(heights, height)
	ews.Set(subID, pubID, heights)
	return heights
}

func (ews *EventWarehouseState) ToBytes() ([]byte, error) {
	var bs []byte
	bf := bytes.NewBuffer(bs)
	if err := gob.NewEncoder(bf).Encode(ews.state); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
