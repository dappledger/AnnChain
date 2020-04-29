package evm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/dappledger/AnnChain/eth/ethdb"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/types"
	"go.uber.org/zap"
)

var (
	KvHistoryPrefix = []byte("kvh_")
	KvSizePrefix    = []byte("_size_")
	kvIndexPrefix = []byte("_ff_")
)

const (
	PageNumLen = 4
	PageSizeLen = 4
)

func makeKey(key []byte, suffix[] byte) []byte {
	buf:= bytes.NewBuffer(nil)
	buf.Write(KvHistoryPrefix)
	buf.Write(key)
	buf.Write([]byte("_ff_"))
	buf.Write(suffix)
	return buf.Bytes()
}

func makeKeySizeKey(key []byte) []byte {
	buf:= bytes.NewBuffer(nil)
	buf.Write(KvHistoryPrefix)
	buf.Write(key)
	buf.Write([]byte("_fd_"))
	buf.Write( KvSizePrefix)
	return buf.Bytes()
}

func putUint32(i uint32) []byte {
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, i)
	return index
}


type KeyValueHistoryManager struct {
	db ethdb.Database
	mu sync.RWMutex
}

func (m *KeyValueHistoryManager) Close() {
	m.db.Close()
}

func NewKeyValueHistoryManager(db ethdb.Database) *KeyValueHistoryManager {
	return &KeyValueHistoryManager{db: db}
}


func (m *KeyValueHistoryManager) SaveKeyHistory(kvs types.KeyValueHistories) error {
	if len(kvs) ==0 {
		return nil
	}
	batch := m.NewBatch()
	return batch.SaveKeyHistory(kvs)
}


func (m *KeyValueHistoryManager) NewBatch() *kvBatch {
	k := &kvBatch{
		batch:                  m.db.NewBatch(),
		keys:                   make(map[string]uint32),
		KeyValueHistoryManager: m,
	}
	return k
}

func (m *KeyValueHistoryManager) GetKeyHistorySize(key []byte) (uint32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getKeyHistorySize(key)
}

func (m *KeyValueHistoryManager) getKeyHistorySize(key []byte) (uint32, error) {
	k := makeKeySizeKey(key)
	data, err := m.db.Get(k)
	if err != nil {
		return 0, err
	}
	if len(data) < 4 {
		return 0, fmt.Errorf("should be 4 byte %v", data)
	}
	return binary.BigEndian.Uint32(data), nil
}

func (m *KeyValueHistoryManager) Get(key []byte, index uint32) (history *types.ValueUpdateHistory, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.get(key, index)
}

func (m *KeyValueHistoryManager) get(key []byte, index uint32) (*types.ValueUpdateHistory, error) {
	k := makeKey(key, putUint32(index))
	data, err := m.db.Get(k)
	if err != nil {
		return nil, err
	}
	history := &types.ValueUpdateHistory{}
	err = rlp.DecodeBytes(data, history)
	if err != nil {
		return nil, err
	}
	return history, nil
}


func (m *KeyValueHistoryManager) Query(key []byte, pageNo uint32, pageSize uint32) (histories []*types.ValueUpdateHistory, total uint32, err error) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize >20 {
		pageSize = 20
	}
	total, err = m.GetKeyHistorySize(key)
	if err != nil {
		log.Infof("get size err %v", err)
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	var from ,to uint32
	if total < (pageNo-1)*pageSize {
		return nil, total, nil
	}
	from = total - (pageNo-1)*pageSize

	if from >= pageSize {
		to  = from - pageSize
	}else {
		pageSize = from
	}

	var kvs []*ethdb.KVResult
	kvs, err = m.db.GetWithPrefix(makeKey(key, nil), makeKey(key, putUint32(to)), pageSize, 0)
	if err != nil {
		log.Infof("get key history err for key %v   to %v pageSize %v  err %v",string(key),to,pageSize,err)
		return nil, total, err
	}
	for i := len(kvs) - 1; i >= 0; i-- {
		kv := kvs[i]
		history := &types.ValueUpdateHistory{}
		err = rlp.DecodeBytes(kv.V, &history)
		if err != nil {
			log.Infof("decode rlp  err  %v  to %v pageSize %v  err %v %s %s",string(key),to,pageSize,err, i,string(kv.K), string(kv.V))
			return nil, total, err
		}
		histories = append(histories, history)
	}
	return
}


type kvBatch struct {
	batch                  ethdb.Batch
	keys                   map[string]uint32
	KeyValueHistoryManager *KeyValueHistoryManager
}

func (batch *kvBatch) SaveKeyHistory(kvs types.KeyValueHistories) error {
	batch.KeyValueHistoryManager.mu.Lock()
	defer batch.KeyValueHistoryManager.mu.Unlock()
	return batch.saveKeyHistory(kvs)
}

func (batch *kvBatch) saveKeyHistory(kvs types.KeyValueHistories) error {
	log.Infof("save key histories %s ",kvs.String())
	for i := range kvs {
		kv := kvs[i]
		err := batch.put(kv.Key, kv.ValueUpdateHistory)
		if err != nil {
			return err
		}
	}
	for key, size := range batch.keys {
		err := batch.setKeyHistorySize([]byte(key), size)
		if err != nil {
			log.Warnf("write err %v %v %v", key, size, err)
			return err
		}
	}
	return batch.batch.Write()
}


func (batch *kvBatch) put(key []byte, history *types.ValueUpdateHistory) error {
	var size uint32
	var ok bool
	var err error
	if size, ok = batch.keys[string(key)]; !ok {
		size, err = batch.KeyValueHistoryManager.getKeyHistorySize(key)
		if err != nil {
			//check not found
			log.Info("read size err", zap.Error(err))
		}
	}
	err = batch.putHistory(key, size, history)
	if err != nil {
		log.Warnf("write err %v %v %v %v", key, size, err, history)
		return err
	}
	batch.keys[string(key)] = size + 1
	return nil
}

func (batch *kvBatch) Put(key []byte, history *types.ValueUpdateHistory) error {
	batch.KeyValueHistoryManager.mu.Lock()
	defer batch.KeyValueHistoryManager.mu.Unlock()
	return batch.put(key, history)
}

func (batch *kvBatch) setKeyHistorySize(key []byte, size uint32) error {
	k := makeKeySizeKey(key)
	return batch.batch.Put(k, putUint32(size))
}

func (batch *kvBatch) putHistory(key []byte, index uint32, history *types.ValueUpdateHistory) error {
	k := makeKey(key, putUint32(index))
	historyBytes, err := rlp.EncodeToBytes(history)
	if err != nil {
		return err
	}
	return batch.batch.Put(k, historyBytes)
}