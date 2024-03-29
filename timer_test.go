package kebench_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type stimer struct {
	sync.Once
	now int64
}

func (t *stimer) run() {
	t.Do(func() {
		atomic.StoreInt64(&t.now, time.Now().UnixNano())
		go func() {
			for now := range time.Tick(time.Microsecond) {
				atomic.StoreInt64(&t.now, now.UnixNano())
			}
		}()
	})
}

func (t *stimer) Now() time.Time {
	return time.Unix(0, atomic.LoadInt64(&t.now))
}

func TestTimer(t *testing.T) {
	t.Run("Timer-1-1", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			fmt.Println(time.Now())
			time.Sleep(time.Second)
		}
	})
	t.Run("Timer-2-1", func(t *testing.T) {
		st := &stimer{}
		st.run()
		for i := 0; i < 10; i++ {
			fmt.Println(st.Now())
			time.Sleep(time.Second)
		}
	})
}

func BenchmarkTimer(b *testing.B) {
	b.Run("Timer-1-1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Now()
		}
	})
	b.Run("Timer-2-1", func(b *testing.B) {
		t := &stimer{}
		t.run()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.Now()
		}
	})
}
