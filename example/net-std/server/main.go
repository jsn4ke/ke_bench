package main

import (
	"flag"
	"fmt"
	"io"
	"net"

	"net/http"
	_ "net/http/pprof"

	kebench "github.com/jsn4ke/ke_bench"
	netstd "github.com/jsn4ke/ke_bench/example/net-std"
)

var (
	addr  string
	ctype int
)

func initFlag() {
	flag.StringVar(&addr, "s", ":9999", "server address")
	flag.IntVar(&ctype, "t", 1, "codec type")
}

func main() {
	initFlag()
	flag.Parse()
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	ln, err := net.Listen("tcp", addr)
	if nil != err {
		panic(err)
	}
	var (
		create func(io.ReadWriter) netstd.Codec
	)
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
	for {
		conn, err := ln.Accept()
		if nil != err {
			panic(err)
		}
		go func() {
			defer conn.Close()
			codec := create(conn)
			for {
				req, err := codec.Decode()
				if nil != err {
					break
				}
				resp := kebench.ProcessMessage(req)
				err = codec.Encode(resp)
				if nil != err {
					break
				}
			}
		}()
	}
}
