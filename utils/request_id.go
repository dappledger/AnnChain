package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var RequestIdHeader = "X-B3-TraceId"
var requestId uint32

type TraceID struct {
	High, Low uint64
}

func NewTraceId(timestamp time.Time) TraceID {
	t := TraceID{}
	t.Low = randomNumber()
	t.High = uint64(uint32(timestamp.Unix()))<<16 + uint64(atomic.AddUint32(&requestId, 1))
	return t
}

// Timestamp extracts the time part of the TraceID.
func (id TraceID) Timestamp() time.Time {
	unixSecs := id.High >> 16
	return time.Unix(int64(unixSecs), 0)
}

func (t TraceID) String() string {
	if t.High == 0 {
		return fmt.Sprintf("%x", t.Low)
	}
	return fmt.Sprintf("%x%016x", t.High, t.Low)
}

var randomNumber func() uint64

func init() {
	seedGenerator := NewRand(time.Now().UnixNano())
	pool := sync.Pool{
		New: func() interface{} {
			return rand.NewSource(seedGenerator.Int63())
		},
	}
	randomNumber = func() uint64 {
		generator := pool.Get().(rand.Source)
		number := uint64(generator.Int63())
		pool.Put(generator)
		return number
	}
}

// lockedSource allows a random number generator to be used by multiple goroutines concurrently.
// The code is very similar to math/rand.lockedSource, which is unfortunately not exposed.
type lockedSource struct {
	mut sync.Mutex
	src rand.Source
}

// NewRand returns a rand.Rand that is threadsafe.
func NewRand(seed int64) *rand.Rand {
	return rand.New(&lockedSource{src: rand.NewSource(seed)})
}

func (r *lockedSource) Int63() (n int64) {
	r.mut.Lock()
	n = r.src.Int63()
	r.mut.Unlock()
	return
}

// Seed implements Seed() of Source
func (r *lockedSource) Seed(seed int64) {
	r.mut.Lock()
	r.src.Seed(seed)
	r.mut.Unlock()
}

// TraceIDFromString creates a TraceID from a hexadecimal string
func TraceIDFromString(s string) (TraceID, error) {
	var hi, lo uint64
	var err error
	if len(s) > 32 || len(s) <= 16 {
		return TraceID{}, fmt.Errorf("TraceID cannot be longer than 32 hex characters or shorter than 16 hex: %s", s)
	} else {
		hiLen := len(s) - 16
		if hi, err = strconv.ParseUint(s[0:hiLen], 16, 64); err != nil {
			return TraceID{}, err
		}
		if lo, err = strconv.ParseUint(s[hiLen:], 16, 64); err != nil {
			return TraceID{}, err
		}
	}
	return TraceID{High: hi, Low: lo}, nil
}
