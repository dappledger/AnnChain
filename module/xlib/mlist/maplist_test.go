package mlist

import (
	"strconv"
	"testing"
)

type TestElemData struct {
	key   int
	value int
}

func (td *TestElemData) Init(key, value int) {
	td.key = key
	td.value = value
}

func (td *TestElemData) Key() string {
	return strconv.Itoa(td.key)
}

func TestMapList(t *testing.T) {
	l := NewMapList()
	for i := 0; i < 100; i++ {
		td := &TestElemData{}
		td.Init(i, i*2)
		l.Set(td.Key(), td)
	}

	var i int
	l.Exec(func(key string, e interface{}) {
		if v, ok := e.(*TestElemData); ok {
			if i != v.key || v.value != i*2 {
				t.Error("wrong index", i, v.key, v.value)
			}
		}
		i++
	})
	if i != 100 || l.Len() != i {
		t.Error("wrong size", l.Len())
	}

	for i := 0; i < 10; i++ {
		l.Del(strconv.Itoa(i))
	}

	for i := 99; i >= 80; i-- {
		l.Del(strconv.Itoa(i))
	}

	for i := 29; i < 39; i++ {
		l.Del(strconv.Itoa(i))
	}

	i = 0
	l.Exec(func(key string, e interface{}) {
		if v, ok := e.(*TestElemData); ok {
			if v.key < 10 || v.key >= 80 || (v.key >= 30 && v.key < 39) {
				t.Error("wrong index", i, v.key, v.value)
			}
		}
		i++
	})
	if i != 100-40 || l.Len() != i {
		t.Error("wrong size", l.Len())
	}

	for i := 0; i < 100; i++ {
		l.Del(strconv.Itoa(i))
	}
	if l.Len() != 0 {
		t.Error("wrong size", l.Len())
	}
}

const CountElemData = 10000

var vl = NewMapList()
var ar = make([]interface{}, CountElemData)

var TestFunc = func(key string, data interface{}) {
	if v, ok := data.(*TestElemData); ok {
		_ = v
	}
}

func ExecArray(exec func(key string, data interface{})) {
	for i := range ar {
		exec("", ar[i])
	}
}

func init() {
	for i := 0; i < CountElemData; i++ {
		td := &TestElemData{}
		td.Init(i, i*2)
		vl.Set(td.Key(), td)
		ar[i] = td
	}
}

func BenchmarkMapListSet(b *testing.B) {
	l := NewMapList()
	for i := 0; i < b.N; i++ {
		td := &TestElemData{}
		td.Init(i, i*2)
		l.Set(td.Key(), td)
	}
}

func BenchmarkArraySet(b *testing.B) {
	a := make([]*TestElemData, 0)
	keyMap := make(map[string]interface{})
	for i := 0; i < b.N; i++ {
		td := &TestElemData{}
		td.Init(i, i*2)
		keyMap[td.Key()] = td
		a = append(a, td)
	}
}

func BenchmarkMapListExec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vl.Exec(TestFunc)
	}
}

func BenchmarkArrayExec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExecArray(TestFunc)
	}
}
