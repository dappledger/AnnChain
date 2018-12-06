package session

import (
	"testing"
	"time"
)

func TestBench(t *testing.T) {

	s := NewSession(10, 2)

	s.SetSession("name", "fhy")

	for {
		v := s.GetSession("name")
		if v == nil {
			s.Close()
			return
		} else {
			t.Log(v)
		}
		time.Sleep(time.Second)
	}
}
