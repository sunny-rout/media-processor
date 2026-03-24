package worker

import "errors"

var ErrPoolFull = errors.New("server is at capacity, too many concurrent streams")

type Pool struct {
	sem chan struct{}
}

func NewPool(maxConcurrent int) *Pool {
	return &Pool{sem: make(chan struct{}, maxConcurrent)}
}

func (p *Pool) Acquire() error {
	select {
	case p.sem <- struct{}{}:
		return nil
	default:
		return ErrPoolFull
	}
}

func (p *Pool) Release() {
	<-p.sem
}

func (p *Pool) Active() int {
	return len(p.sem)
}

func (p *Pool) Capacity() int {
	return cap(p.sem)
}
