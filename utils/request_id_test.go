package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTraceId (t *testing.T) {
	begin := time.Now()
	traceIdStr := NewTraceId(begin).String()
	traceId,err := TraceIDFromString(traceIdStr)
	assert.NoError(t,err)
	assert.Equal(t,traceId.Timestamp().Unix(),begin.Unix(),)

}