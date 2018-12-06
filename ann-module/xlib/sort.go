package xlib

// default serial model
type SortKey interface {
	String() string
}

type Sortable interface {
	Key() SortKey
	Less(Sortable) bool
}

type SortItfc interface {
	Init(maxSize int) // <=0 means no limit
	Put(data Sortable) bool
	Drop(key string)
	Get(string) Sortable
	Exec(func(data Sortable) bool)
	ChangeData(data Sortable) // modify exist element's value

	Sort() // active call, for some structs just a return
}

type SORTER_TYPE int

const (
	SORTER_TYPE_LIST SORTER_TYPE = iota + 1
)

func NewSorter(size int, tp SORTER_TYPE) SortItfc {
	switch tp {
	case SORTER_TYPE_LIST:
		var sl SortList
		sl.Init(size)
		return &sl
	}

	return nil
}
