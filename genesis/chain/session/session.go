package session

import (
	"time"
)

type Session struct {
	checkTime  int64
	expireTime int64
	exitChan   chan bool
	mMap       *Map
}

type SessionValue struct {
	CreateTime int64
	Value      interface{}
}

func NewSession(expireTime, checkTime int64) *Session {

	s := &Session{
		mMap:       new(Map),
		expireTime: expireTime,
		checkTime:  checkTime,
		exitChan:   make(chan bool, 0),
	}

	go func() {
		chanCheck := time.NewTicker(time.Duration(s.checkTime))
		defer chanCheck.Stop()
		for {
			select {
			case <-chanCheck.C:
				s.mMap.RLockRange(func(key interface{}, value interface{}) {
					if time.Now().Unix()-value.(SessionValue).CreateTime >= s.expireTime {
						s.mMap.UnsafeDel(key)
					}
				})
			case <-s.exitChan:
				return
			}

		}
	}()

	return s
}

func (s *Session) SetSession(key, value interface{}) {
	s.mMap.Set(key, SessionValue{
		Value:      value,
		CreateTime: time.Now().Unix(),
	})
}

func (s *Session) GetSessionAndDelete(key interface{}) interface{} {
	sv := s.mMap.GetAndDel(key)
	if sv == nil {
		return nil
	}
	return sv.(SessionValue).Value
}

func (s *Session) GetSession(key interface{}) interface{} {
	sv := s.mMap.Get(key)
	if sv == nil {
		return nil
	}
	return sv.(SessionValue).Value
}

func (s *Session) Close() {
	close(s.exitChan)
}
