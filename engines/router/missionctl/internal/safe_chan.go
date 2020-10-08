package internal

import (
	"sync"
)

// SafeChan is a wrapper around golang `chan` that solves the problem of a
// graceful closing of the channel, that is being written by several goroutines.
// @See: https://gist.github.com/leolara/f6fb5dfc04d64947487f16764d6b37b6
type SafeChan struct {
	ch chan interface{}

	closingCh      chan interface{}
	writersWG      sync.WaitGroup
	writersWGMutex sync.Mutex
}

// NewSafeChan creates a SafeChan
func NewSafeChan(buf ...int) *SafeChan {
	var ch chan interface{}
	if len(buf) > 0 {
		ch = make(chan interface{}, buf[0])
	} else {
		ch = make(chan interface{})
	}

	return &SafeChan{
		ch:        ch,
		closingCh: make(chan interface{}),
	}
}

// Read returns the channel to write
func (sc *SafeChan) Read() <-chan interface{} {
	return sc.ch
}

// Write writes data into the channel
func (sc *SafeChan) Write(data interface{}) {
	sc.writersWGMutex.Lock()
	sc.writersWG.Add(1)
	sc.writersWGMutex.Unlock()
	defer sc.writersWG.Done()

	select {
	case <-sc.closingCh:
		return
	default:
	}

	select {
	case <-sc.closingCh:
	case sc.ch <- data:
	}
}

// WriteAsync writes into the channel in a different goroutine
func (sc *SafeChan) WriteAsync(data interface{}) {
	go sc.Write(data)
}

// Close closes channel, draining any blocked writes
func (sc *SafeChan) Close() {
	close(sc.closingCh)

	go func() {
		for range sc.ch {
		}
	}()

	sc.writersWGMutex.Lock()
	sc.writersWG.Wait()
	sc.writersWGMutex.Unlock()

	close(sc.ch)
}

// CloseWithoutDraining closes channel, without draining any pending writes,
// this method will block until all writes have been unblocked by reads
func (sc *SafeChan) CloseWithoutDraining() {
	close(sc.closingCh)

	sc.writersWGMutex.Lock()
	sc.writersWG.Wait()
	sc.writersWGMutex.Unlock()

	close(sc.ch)
}
