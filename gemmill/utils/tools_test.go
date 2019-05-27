package utils

import (
	"testing"
)

func TestSortInt64Slc(t *testing.T) {
	var slc = []int64{1, 129, 20, 4, 45, 66}
	Int64Slice(slc).Sort()
	pre := slc[0]
	for i := range slc {
		if pre > slc[i] {
			t.Error("sort err")
			return
		}
		pre = slc[i]
	}
}
