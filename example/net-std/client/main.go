package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"time"
	"unsafe"

	kebench "github.com/jsn4ke/ke_bench"
	netstd "github.com/jsn4ke/ke_bench/example/net-std"
)

var (
	addr        string
	concurrency int
	total       int
	bodySize    int
	ctype       int
)

func initFlag() {
	flag.StringVar(&addr, "s", ":9999", "server address")
	flag.IntVar(&concurrency, "c", 1, "concurrency")
	flag.IntVar(&total, "n", 1, "total")
	flag.IntVar(&bodySize, "b", 1024, "body size")
	flag.IntVar(&ctype, "t", 1, "codec type")
}

type clientUnit struct {
	pool *kebench.ConnectionPool[net.Conn]
}

var (
	noconn = errors.New("no connection")
	create func(io.ReadWriter) netstd.Codec
)

func (c *clientUnit) WarmUp() error {
	conn, ok := c.pool.Get()
	if !ok {
		return noconn
	}
	var err error

	defer func() {
		c.pool.Push(conn, err)
	}()

	codec := netstd.NewEchoCodec(conn)
	err = codec.Encode(&kebench.BenchMessage{})
	if nil != err {
		return err
	}

	_, err = codec.Decode()
	return err
}

func (c *clientUnit) Run() error {
	conn, ok := c.pool.Get()
	if !ok {
		return noconn
	}
	var err error

	defer func() {
		c.pool.Push(conn, err)
	}()

	codec := netstd.NewEchoCodec(conn)
	body := make([]byte, bodySize)

	err = codec.Encode(&kebench.BenchMessage{Msg: unsafe.String(&body[0], len(body))})
	if nil != err {
		return err
	}

	_, err = codec.Decode()
	return err
}

func (c *clientUnit) Begin() error {
	conn, ok := c.pool.Get()
	if !ok {
		return noconn
	}
	var err error

	defer func() {
		c.pool.Push(conn, err)
	}()
	codec := netstd.NewEchoCodec(conn)
	err = codec.Encode(&kebench.BenchMessage{})
	if nil != err {
		return err
	}

	_, err = codec.Decode()
	return err
}

func (c *clientUnit) End() error {
	conn, ok := c.pool.Get()
	if !ok {
		return noconn
	}
	var err error

	defer func() {
		c.pool.Push(conn, err)
	}()

	codec := netstd.NewEchoCodec(conn)
	err = codec.Encode(&kebench.BenchMessage{})
	if nil != err {
		return err
	}

	_, err = codec.Decode()
	return err
}

func main() {
	initFlag()
	flag.Parse()

	fmt.Println("server address:", addr, "concurrency:", concurrency,
		"total:", total, "body size:", bodySize, "codec type:", ctype)

	switch ctype {
	case 1:
		create = netstd.NewEchoCodec
	case 2:
		create = netstd.NewEchoCodec2
	case 3:
		create = netstd.NewEchoCodec3
	case 4:
		create = netstd.NewEchoCodec4
	default:
		fmt.Println("invalid codec type", ctype)
		return
	}

	runner := kebench.NewRunner(time.Now)

	unit := new(clientUnit)
	atleast := concurrency
	if atleast < 1024 {
		atleast = 1024
	}
	unit.pool = kebench.NewConnectionPool[net.Conn](func() (net.Conn, bool) {
		conn, err := net.Dial("tcp", addr)
		if nil != err {
			return nil, false
		}
		return conn, true
	}, func(conn net.Conn) {
		conn.Close()
	}, atleast)

	runner.Run(context.Background(), unit, concurrency, int64(total))
}
