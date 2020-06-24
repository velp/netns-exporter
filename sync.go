package main

import (
	"sync"
)

type LimitedWaitGroup struct {
	wg              sync.WaitGroup
	mu              *sync.Mutex
	cond            *sync.Cond
	limit, copasity int
}

func NewLimitedWaitGroup(limit int) *LimitedWaitGroup {
	mu := new(sync.Mutex)
	return &LimitedWaitGroup{
		mu:       mu,
		cond:     sync.NewCond(mu),
		limit:    limit,
		copasity: limit,
	}
}

func (lg *LimitedWaitGroup) Add(delta int) {
	lg.mu.Lock()
	defer lg.mu.Unlock()
	if delta > lg.limit {
		panic(`LimitedWaitGroup: delta must not exceed limit`)
	}
	for lg.copasity < 1 {
		lg.cond.Wait()
	}
	lg.copasity -= delta
	lg.wg.Add(delta)
}

func (lg *LimitedWaitGroup) Done() {
	lg.mu.Lock()
	defer lg.mu.Unlock()
	lg.copasity++
	lg.cond.Signal()
	lg.wg.Done()
}

func (lg *LimitedWaitGroup) Wait() {
	lg.wg.Wait()
}
