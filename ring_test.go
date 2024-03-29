package kebench_test

import (
	"testing"

	kebench "github.com/jsn4ke/ke_bench"
)

func TestRingPush(t *testing.T) {
	t.Log("TestRingGet")

	r := kebench.NewRing[int](10)
	for i := 0; i < 10; i++ {
		if !r.Push(i) {
			t.Error("push failed")
		}
	}
	for i := 0; i < 10; i++ {
		if r.Push(i) {
			t.Error("push need failed")
		}
	}
}

func TestRingGet(t *testing.T) {
	t.Log("TestRingGet")

	r := kebench.NewRing[int](10)
	for i := 0; i < 10; i++ {
		r.Push(i)
	}
	for i := 0; i < 10; i++ {
		if _, ok := r.Get(); !ok {
			t.Error("get failed")
		}
	}
	for i := 0; i < 10; i++ {
		if _, ok := r.Get(); ok {
			t.Error("get need failed")
		}
	}
}
