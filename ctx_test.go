package kebench_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	go func() {
		tk := time.NewTicker(time.Second)
		for {
			select {

			case <-tk.C:
				fmt.Println(time.Now().String())
			}
		}
	}()
	<-ctx.Done()

	fmt.Println("over")
}

func replaceNewlines(in []byte) []byte {
	for i := 0; i < len(in)-1; i++ {
		if in[i] == '\n' {
			in[i] = '\t'
		}
	}
	return in
}

func TestSpace(t *testing.T) {
	// fmt.Printf("hello\tworld\n")
	// fmt.Printf("hello\nworld\n")
	// fmt.Printf("hello world\n")
	bs := []byte("hello\tworld!\nover\n")
	fmt.Println(string(bs))
	bs2 := bytes.Replace(bs, []byte("\n"), []byte("\t"), -1)
	fmt.Println(string(bs))
	fmt.Println(string(bs2))
	bs3 := replaceNewlines(bs)
	fmt.Println(string(bs3))
	fmt.Println("over")
}

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ctx, _ = context.WithTimeout(ctx, time.Second)
		<-ctx.Done()
		fmt.Println("over 1s", time.Now())
	}()
	go func() {
		ctx, _ = context.WithTimeout(ctx, time.Second*2)
		<-ctx.Done()
		fmt.Println("over 2s", time.Now())
	}()
	go func() {
		ctx, _ = context.WithTimeout(ctx, time.Second*3)
		<-ctx.Done()
		fmt.Println("over 3s", time.Now())
	}()
	go func() {
		<-time.After(time.Second * 4)
		fmt.Println("over 4s", time.Now())
		cancel()
	}()
	<-ctx.Done()
	fmt.Println("over", time.Now())
}
