package state

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestTPS(t *testing.T) {
	tps := NewTPSCalculator(5)
	for i := 0; i < 20; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.Int()%1000))
		tps.AddRecord(uint32(rand.Int()%1000 + 1000))
		fmt.Println("tps now:", tps.TPS())
	}
}
