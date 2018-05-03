package mlist

type Element struct {
	key   string
	value interface{}
	next  *Element
	prev  *Element
}

func (e *Element) clear() {
	e.next = nil
	e.prev = nil
}

type MapList struct {
	head  *Element
	tail  *Element
	keyMp map[string]*Element
}

func NewMapList() *MapList {
	var m MapList
	m.Init()
	return &m
}

func (l *MapList) Init() {
	l.keyMp = make(map[string]*Element)
}

func (l *MapList) Len() int {
	return len(l.keyMp)
}

// Set if the key can't find in the list
// then insert the value into the end of the list,
// or modify the value of the key
func (l *MapList) Set(key string, value interface{}) {
	if v, ok := l.keyMp[key]; ok {
		v.value = value
		return
	}
	e := &Element{
		key: key,
	}
	e.value = value
	if l.tail != nil {
		l.tail.next = e
	}
	e.prev = l.tail
	l.tail = e
	if l.head == nil {
		l.head = e
	}
	l.keyMp[key] = e
}

// Get get the key in the list
func (l *MapList) Get(key string) (interface{}, bool) {
	if v, ok := l.keyMp[key]; ok {
		return v.value, ok
	}
	return nil, false
}

// Del delete the key in the list
func (l *MapList) Del(key string) {
	if v, ok := l.keyMp[key]; ok {
		if l.head == v {
			l.head = v.next
		}
		if l.tail == v {
			l.tail = v.prev
		}
		if v.prev != nil {
			v.prev.next = v.next
		}
		if v.next != nil {
			v.next.prev = v.prev
		}
		v.clear()
		delete(l.keyMp, key)
	}
}

// Has check whether the key is in the list
func (l *MapList) Has(key string) bool {
	_, ok := l.keyMp[key]
	return ok
}

// Exec go through the list
func (l *MapList) Exec(exec func(string, interface{})) {
	var p *Element
	for p = l.head; p != nil; p = p.next {
		exec(p.key, p.value)
	}
}

// Exec go through the list
func (l *MapList) ExecBreak(exec func(string, interface{}) bool) bool {
	var p *Element
	for p = l.head; p != nil; p = p.next {
		if !exec(p.key, p.value) {
			return false
		}
	}
	return true
}

func (l *MapList) Reset() {
	l.Init()
	l.head = nil
	l.tail = nil
}
