package internal_test

import (
	"math/big"
	"sync"
	"testing"

	"github.com/gojek/turing/engines/router/missionctl/internal"
)

func TestSafeChan_Simple(t *testing.T) {
	consumer := func(pub *internal.SafeChan, n int, wg *sync.WaitGroup, result chan *big.Int) {
		ch := pub.Read()

		acc := big.NewInt(0)

		for i := 0; i < n; i++ {
			val := (<-ch).(big.Int)
			acc.Add(acc, &val)
		}

		wg.Done()
		result <- acc
	}

	producer := func(safeChan *internal.SafeChan, n int, wg *sync.WaitGroup) {
		for i := 0; i < n; i++ {
			safeChan.Write(*big.NewInt(int64(i)))
		}
		wg.Done()
	}

	precalc := func(n int) *big.Int {
		acc := big.NewInt(0)
		for i := 0; i < n; i++ {
			acc.Add(acc, big.NewInt(int64(i)))
		}

		return acc
	}

	p := internal.NewSafeChan()
	var wg sync.WaitGroup
	resultCh := make(chan *big.Int)

	wg.Add(2)
	go consumer(p, 100, &wg, resultCh)
	go producer(p, 100, &wg)
	wg.Wait()
	p.CloseWithoutDraining()

	result := <-resultCh

	if result.Cmp(precalc(100)) != 0 {
		t.Error("wrong result")
	}
}

func TestSafeChan_Intermediate(t *testing.T) {
	consumer := func(safeChan *internal.SafeChan, n int, wg *sync.WaitGroup, result chan *big.Int) {
		ch := safeChan.Read()

		acc := big.NewInt(0)

		for i := 0; i < n; i++ {
			val := (<-ch).(big.Int)
			acc.Add(acc, &val)
		}

		wg.Done()
		result <- acc
	}

	producer := func(safeChan *internal.SafeChan, n int, wg *sync.WaitGroup) {
		for i := 0; i < n; i++ {
			safeChan.Write(*big.NewInt(int64(i)))
		}
		wg.Done()
	}

	p := internal.NewSafeChan()
	var wg sync.WaitGroup
	resultCh := make(chan *big.Int)

	wg.Add(3)
	go consumer(p, 100, &wg, resultCh)
	go producer(p, 100, &wg)
	go producer(p, 100, &wg)

	<-resultCh
	p.Close()

	wg.Wait()
}
