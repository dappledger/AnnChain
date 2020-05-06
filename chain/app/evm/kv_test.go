package evm

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dappledger/AnnChain/eth/ethdb"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/types"
	"github.com/stretchr/testify/assert"
)

func TestLevelDb(t *testing.T) {

	//key 0-1 00
	//pageNo:10, pageSize:100
	//iterator  sort
	//key_size  kry_0 ....key_100
	//key_90  key_95
	db, err := ethdb.NewLDBDatabase("_testdb", 0, 0)
	assert.NoError(t, err)
	defer func() {
		time.Sleep(time.Second)
		os.RemoveAll("_testdb")
	}()
	m := NewKeyValueHistoryManager(db)
	defer m.Close()
	var histories types.KeyValueHistories
	genValue := func(val string, suffix string) []byte {
		return []byte(val + "_" + suffix)
	}
	for j := 0; j < 10; j++ {
		for i := 0; i < 100; i++ {
			history := &types.KeyValueHistory{
				Key: []byte(fmt.Sprintf("key_%d", j)),
				ValueUpdateHistory: &types.ValueUpdateHistory{
					TimeStamp:   uint64(time.Now().Unix()),
					Value:       genValue("value", fmt.Sprintf("%d_%d", i, j)),
					TxHash:      []byte("0x5555"),
					BlockHeight: 10,
					TxIndex:12,
				},
			}
			histories = append(histories, history)
		}

	}

	err = m.SaveKeyHistory(histories)
	assert.NoError(t, err)
	result, total, err := m.Query([]byte(fmt.Sprintf("key_%d", 4)), 4, 6)
	assert.NoError(t, err)
	assert.Equal(t, 100, int(total))
	assert.NotNil(t, result)
	from := 100 - 3*6
	for i, v := range result {
		log.Debugf("result %d %v ", i, v, string(v.Value))
		assert.Equal(t, v.Value, genValue("value", fmt.Sprintf("%d_%d", from-i, 4)))
	}
}
