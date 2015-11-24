package main

import (
	"sync/atomic"
)

var global_id uint64 = 0

func SrsGenerateId() uint64 {
	return atomic.AddUint64(&global_id, 1)
}
