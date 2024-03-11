package analytics

import "sync/atomic"

type TransactionCounter struct {
	val int32
}

func (c *TransactionCounter) Inc() int32 {
	return atomic.AddInt32(&c.val, 1)
}

func (c *TransactionCounter) Get() int32 {
	return atomic.LoadInt32(&c.val)
}
