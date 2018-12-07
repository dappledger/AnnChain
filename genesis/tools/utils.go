package tools

import (
	"math/big"
	"unsafe"

	"github.com/dappledger/AnnChain/genesis/types"
)

func CheckItfcNil(itfc interface{}) bool {
	d := (*struct {
		itab uintptr
		data uintptr
	})(unsafe.Pointer(&itfc))
	return d.data == 0
}

func IsSameWeek(time1, time2 uint64) bool {
	if time1 < types.FIRST_MONDAY || time2 < types.FIRST_MONDAY {
		return false
	}
	return (time1-types.FIRST_MONDAY)/7 == (time2-types.FIRST_MONDAY)/7
}

// a * b / c
func BigDivide(a, b, c *big.Int) *big.Int {
	if c.Cmp(types.BIG_INT0) == 0 {
		return types.BIG_INT0
	}
	mul := a.Mul(a, b)
	return mul.Div(mul, c)
}

// round_down
func OriDivide(a, b, c uint64) uint64 {
	if c == 0 {
		return 0
	}
	return uint64(float64(a) / float64(b) * float64(c))
}
