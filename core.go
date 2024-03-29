package kebench

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Unit interface {
	WarmUp() error
	Run() error
	Begin() error
	End() error
}

type Runner struct {
	Now     func() time.Time
	records Records
}

func NewRunner(now func() time.Time) *Runner {
	return &Runner{
		Now: now,
	}
}

type Records struct {
	entry []RecordEntry
}

type RecordEntry struct {
	Cost int64
	Err  error
}

type Handler func() error

func (r *Runner) Run(ctx context.Context, unit Unit, concurrency int, total int64) error {

	fmt.Println("start warmup")
	// warm up
	r.benching(ctx, unit.WarmUp, concurrency, total)

	fmt.Println("start bench")
	if err := unit.Begin(); nil != err {
		return err
	}
	begin := r.Now()
	// running
	r.benching(ctx, unit.Run, concurrency, total)
	end := r.Now()
	if err := unit.End(); nil != err {
		return err
	}
	cost := end.Sub(begin)
	fmt.Printf("bench cost %v\n", cost)
	r.report(cost)
	return nil
}

func (r *Runner) report(cost time.Duration) {
	var totalCost int64
	var totalErrors int64
	for _, entry := range r.records.entry {
		totalCost += entry.Cost
		if entry.Err != nil {
			totalErrors++
		}
	}
	errorTypes := make(map[string]int)
	for _, entry := range r.records.entry {
		if entry.Err != nil {
			errorTypes[entry.Err.Error()]++
		}
	}

	avgCost := float64(totalCost) / float64(len(r.records.entry))
	errorRate := float64(totalErrors) / float64(len(r.records.entry))

	fmt.Printf("Total Requests: %d\n", len(r.records.entry))
	fmt.Printf("Total Cost: %v\n", cost)
	fmt.Printf("Total Sum : %dns, %v\n", totalCost, time.Duration(totalCost))
	fmt.Printf("Average Cost: %.2f, %v\n", avgCost, time.Duration(avgCost))
	if 0 != len(errorTypes) {
		fmt.Println("Error Types:")
		for err, count := range errorTypes {
			fmt.Printf("%s: %d\n", err, count)
		}
	}
	if 0 != totalErrors {
		fmt.Printf("Error Rate: %.2f%%\n", errorRate*100)
	}
	tps := float64(len(r.records.entry)) / (float64(cost) / float64(time.Second))
	fmt.Printf("TPS: %.2f\n", tps)

	// Calculate the median cost
	sortedCosts := make([]int64, len(r.records.entry))
	for i, entry := range r.records.entry {
		sortedCosts[i] = entry.Cost
	}
	sort.Slice(sortedCosts, func(i, j int) bool {
		return sortedCosts[i] < sortedCosts[j]
	})
	medianCost := sortedCosts[len(sortedCosts)/2]
	fmt.Printf("Median Cost: %d, %v\n", medianCost, time.Duration(medianCost))
	// Calculate the cost at different percentiles
	percentiles := []float64{0.1, 0.3, 0.7, 0.8, 0.9, 0.99}
	for _, percentile := range percentiles {
		index := int(float64(len(sortedCosts)) * percentile)
		cost := sortedCosts[index]
		fmt.Printf("Cost at %.2f%%: %d, %v\n", percentile*100, cost, time.Duration(cost))
	}

}

func (r *Runner) benching(ctx context.Context, handler Handler, concurrency int, total int64) {
	var (
		idx int64
		wg  sync.WaitGroup
	)
	idx = 0
	wg.Add(concurrency)
	r.records.entry = make([]RecordEntry, total)
	for i := 0; i < concurrency; i++ {
		// code for each iteration
		go func() {
			defer wg.Done()
			for {
				i := atomic.AddInt64(&idx, 1)
				if i > total {
					return
				}
				cost, err := r.wrapExec(ctx, handler)
				r.records.entry[i-1] = RecordEntry{
					Cost: cost,
					Err:  err,
				}
			}
		}()
	}
	wg.Wait()
}

var (
	ErrTimeout = errors.New("timeout")
)

func (r *Runner) wrapExec(ctx context.Context, handler Handler) (cost int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var (
		done = make(chan error, 1)
	)
	begin := r.Now()
	go func() {
		err := handler()
		select {
		case done <- err:
		default:
		}
	}()
	select {
	case err = <-done:
	case <-ctx.Done():
		err = ErrTimeout
	}
	cost = r.Now().Sub(begin).Nanoseconds()
	return
}
