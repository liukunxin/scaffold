package runtime

import "sync/atomic"

type Sequencer struct {
	n int64
}

func NewSequencer() *Sequencer {
	return &Sequencer{}
}

func (s *Sequencer) Next() int64 {
	return atomic.AddInt64(&s.n, 1)
}
