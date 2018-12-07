package xlib

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

const TEST_NUM = 2

type KeyType string

func (k KeyType) String() string {
	return string(k)
}

type StSortable struct {
	key   KeyType
	value *big.Int
}

func (s *StSortable) Key() SortKey {
	return s.key
}

func (s *StSortable) Less(st Sortable) bool {
	if d, ok := st.(*StSortable); ok {
		return s.value.Cmp(d.value) > 0
	}

	return false
}

func TestSortList(t *testing.T) {
	var sl SortList
	sl.Init(TEST_NUM)
	vslc := make([]StSortable, TEST_NUM*2)
	var i int64
	for i = 0; i < int64(len(vslc)); i++ {
		vslc[i].key = KeyType(fmt.Sprintf("%v", i))
		rand.Seed(time.Now().UnixNano())
		vslc[i].value = big.NewInt(i + rand.Int63()%500)
		sl.Add(&vslc[i])
		fmt.Println("generate:", vslc[i].key, vslc[i].value, sl.Len())
	}

	testPrintList(&sl)

	fmt.Println("===============================")

	for i := 0; i < TEST_NUM; i++ {
		if sl.Get(vslc[i].Key().String()) != nil {
			if i%2 == 0 {
				vslc[i].value.Add(vslc[i].value, big.NewInt(200))
				sl.ChangeData(&vslc[i])
			} else {
				vslc[i].value.Sub(vslc[i].value, big.NewInt(200))
				sl.ChangeData(&vslc[i])
			}
			fmt.Println("modify:", vslc[i].Key().String(), ",add:", i%2 == 0, ",ori:", vslc[i].value)
		} else {
			fmt.Println("not find:", vslc[i].Key().String())
		}
	}

	testPrintList(&sl)

	fmt.Println("===============================")

	for i := 0; i < TEST_NUM/2; i++ {
		if sl.Get(vslc[i].Key().String()) != nil {
			sl.Drop(vslc[i].Key().String())
			fmt.Println("drop:", vslc[i].Key().String(), ",", sl.Len())
		} else {
			fmt.Println("not find:", vslc[i].Key().String())
		}
	}
	testPrintList(&sl)
	fmt.Println(sl.Len())
}

func testPrintList(sl *SortList) {
	var prev *StSortable
	sl.Exec(func(data Sortable) bool {
		if iv, ok := data.(*StSortable); ok {
			if prev != nil && !prev.Less(data) && prev.value.Cmp(iv.value) != 0 {
				fmt.Println("err sort:", prev.Key(), data.Key())
			}
			fmt.Println(iv.key, iv.value)
			prev = iv
		}
		return true
	})
}
