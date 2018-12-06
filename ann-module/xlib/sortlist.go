package xlib

// not thread-safe

type SortList struct {
	first   *element
	last    *element
	size    int
	limit   int
	dataMap map[string]*element
}

type element struct {
	value Sortable
	prev  *element
	next  *element
}

// ret success num
func (sl *SortList) Add(values ...Sortable) uint {
	var oknum uint
	for i := range values {
		if sl.Put(values[i]) {
			oknum++
		}
	}
	return oknum
}

// private func
// if prev is nil, default means it's first,
// if prev.next is nil, default means it's last
func (sl *SortList) _insertE(prev *element, value Sortable) {
	full := sl.Full()
	newElement := &element{value: value, prev: prev}
	if prev == nil {
		if sl.first != nil {
			newElement.next = sl.first
			sl.first.prev = newElement
		}
		sl.first = newElement
		if sl.last == nil {
			sl.last = newElement
		}
	} else {
		if prev.next == nil {
			sl.last = newElement
		} else {
			prev.next.prev = newElement
		}
		newElement.next = prev.next
		prev.next = newElement
	}
	sl.size++
	sl.dataMap[value.Key().String()] = newElement
	if full {
		sl._popend()
	}
}

func (sl *SortList) _dropE(e *element) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		sl.first = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		sl.last = e.prev
	}
	e.prev = nil
	e.next = nil
	delete(sl.dataMap, e.value.Key().String())
	sl.size--
}

func (sl *SortList) _append(values Sortable) {
	sl._insertE(sl.last, values)
}

func (sl *SortList) _popend() {
	sl._dropE(sl.last)
}

func (sl *SortList) _addfirst(values Sortable) {
	sl._insertE(nil, values)
}

func (sl *SortList) Full() bool {
	return sl.limit > 0 && sl.size >= sl.limit
}

func (sl *SortList) AddToEnd(values Sortable) bool {
	if sl.last != nil {
		if sl.last.value.Less(values) {
			if !sl.Full() {
				sl._append(values)
			}
			return true
		}
	} else {
		sl._insertE(nil, values)
		return true
	}
	return false
}

func (sl *SortList) Init(size int) {
	sl.dataMap = make(map[string]*element)
	sl.limit = size
}

// TODO try not begen with index, just rand it
func (sl *SortList) Put(v Sortable) bool {
	if _, ok := sl.dataMap[v.Key().String()]; ok {
		return false
	}
	if sl.AddToEnd(v) {
		return true
	}
	var prev *element
	for e := sl.first; e != nil; e = e.next {
		if e.value.Less(v) {
			prev = e
			if e.next == nil {
				sl._append(v)
			}
			continue
		}
		sl._insertE(prev, v)
		break
	}
	return true
}

func (sl *SortList) ChangeData(data Sortable) {
	if elm := sl.dataMap[data.Key().String()]; elm != nil {
		var prev *element
		prev = elm.prev
		if prev == nil {
			prev = elm.next
		}
		sl._dropE(elm)
		if prev != nil {
			if data.Less(prev.value) {
				for e := prev; e != nil && data.Less(e.value); e = e.prev {
					prev = e.prev

				}
			} else {
				for e := prev; e != nil && !data.Less(e.value); e = e.next {
					prev = e
				}
				if prev == nil {
					sl._append(data)
					return
				}
			}
		}
		sl._insertE(prev, data)
	}
}

func (sl *SortList) Drop(key string) {
	if e := sl.dataMap[key]; e != nil {
		sl._dropE(e)
	}
}

func (sl *SortList) Get(key string) Sortable {
	if data, ok := sl.dataMap[key]; ok {
		return data.value
	}
	return nil
}

func (sl *SortList) Sort() {}

func (sl *SortList) Exec(exec func(data Sortable) bool) {
	for e := sl.first; e != nil; e = e.next {
		if !exec(e.value) {
			break
		}
	}
}

func (sl *SortList) Len() int {
	return sl.size
}

func (sl *SortList) Reset() {
	sl.first = nil
	sl.last = nil
	sl.size = 0
	sl.limit = 0
	sl.dataMap = make(map[string]*element)
}
