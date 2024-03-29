package kebench_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	kebench "github.com/jsn4ke/ke_bench"
)

func TestConnectionPool(t *testing.T) {
	p := kebench.NewConnectionPool[int](func() (int, bool) {
		return 1, true
	}, func(i int) {}, 10)

	done := make(chan struct{})
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				v, ok := p.Get()
				if !ok {
					continue
				}
				defer p.Push(v, nil)
				if 0 == v {
					panic("xx")
				}
			}
		}()
	}
	go func() {
		time.Sleep(time.Second * 10)
		close(done)
	}()
	wg.Wait()
	fmt.Println("over")
}
