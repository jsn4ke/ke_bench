package netstd

import (
	"bufio"
	"encoding/binary"
	"io"
	"math/bits"
	"sync"
	"unsafe"

	kebench "github.com/jsn4ke/ke_bench"
)

type Codec interface {
	Decode() (*kebench.BenchMessage, error)
	Encode(msg *kebench.BenchMessage) error
}

type EchoCodec struct {
	rw io.ReadWriter
}

func NewEchoCodec(rw io.ReadWriter) Codec {
	return &EchoCodec{rw: rw}
}

func (c *EchoCodec) Decode() (*kebench.BenchMessage, error) {
	var head [4]byte
	_, err := io.ReadFull(c.rw, head[:])
	if nil != err {
		return nil, err
	}
	length := binary.BigEndian.Uint32(head[:])
	if length == 0 {
		return &kebench.BenchMessage{Msg: ""}, nil
	}
	data := make([]byte, length)
	_, err = io.ReadFull(c.rw, data)
	if nil != err {
		return nil, err
	}
	tmp := make([]byte, length)
	copy(tmp, data)
	msg := unsafe.String(&tmp[0], len(tmp))
	return &kebench.BenchMessage{Msg: msg}, nil
}

func (c *EchoCodec) Encode(msg *kebench.BenchMessage) error {
	var (
		body = make([]byte, 4+len(msg.Msg))
	)
	binary.BigEndian.PutUint32(body[:4], uint32(len(msg.Msg)))
	data := unsafe.Slice(unsafe.StringData(msg.Msg), len(msg.Msg))
	copy(body[4:], data)
	_, err := c.rw.Write(body)
	return err
}

type EchoCodec2 struct {
	rw bufio.ReadWriter
}

func NewEchoCodec2(rw io.ReadWriter) Codec {
	return &EchoCodec2{rw: bufio.ReadWriter{Reader: bufio.NewReader(rw), Writer: bufio.NewWriter(rw)}}
}

func (c *EchoCodec2) Decode() (*kebench.BenchMessage, error) {
	head, err := c.rw.Peek(4)
	if nil != err {
		return nil, err
	}
	c.rw.Discard(4)
	length := binary.BigEndian.Uint32(head[:])
	if length == 0 {
		return &kebench.BenchMessage{Msg: ""}, nil
	}
	data := make([]byte, length)
	_, err = io.ReadFull(c.rw, data)
	if nil != err {
		return nil, err
	}
	tmp := make([]byte, length)
	copy(tmp, data)
	msg := unsafe.String(&tmp[0], len(tmp))
	return &kebench.BenchMessage{Msg: msg}, nil
}

func (c *EchoCodec2) Encode(msg *kebench.BenchMessage) error {
	head := [4]byte{}
	binary.BigEndian.PutUint32(head[:], uint32(len(msg.Msg)))
	_, err := c.rw.Write(head[:])
	if nil != err {
		return err
	}
	data := unsafe.Slice(unsafe.StringData(msg.Msg), len(msg.Msg))
	_, err = c.rw.Write(data)
	c.rw.Flush()
	return err
}

type Buffer struct {
	Slice []byte
}

func (b *Buffer) grow(size int) {
	newcap := cap(b.Slice)
	doublecap := newcap + newcap
	if size > doublecap {
		newcap = size
	} else {
		if newcap < 1024 {
			newcap = doublecap
		} else {
			// Check 0 < newcap to detect overflow
			// and prevent an infinite loop.
			for 0 < newcap && newcap < size {
				newcap += newcap / 4
			}
			// Set newcap to the requested size when
			// the newcap calculation overflowed.
			if newcap <= 0 {
				newcap = size
			}
		}
	}
	b.Slice = make([]byte, newcap)
}

func (b *Buffer) Reset(size int) {
	if cap(b.Slice) < size { //需要扩容
		b.grow(size)
	}
	b.Slice = b.Slice[:size]
}

type EchoCodec3 struct {
	rw io.ReadWriter
}

var (
	bufferPool = sync.Pool{
		New: func() any {
			return &Buffer{
				Slice: make([]byte, 256),
			}
		},
	}
)

func NewEchoCodec3(rw io.ReadWriter) Codec {
	return &EchoCodec3{rw: rw}
}

func (c *EchoCodec3) Decode() (*kebench.BenchMessage, error) {
	bf := bufferPool.Get().(*Buffer)
	defer func() {
		bufferPool.Put(bf)
	}()

	bf.Reset(4)
	_, err := io.ReadFull(c.rw, bf.Slice)
	if nil != err {
		return nil, err
	}
	length := binary.BigEndian.Uint32(bf.Slice)
	if length == 0 {
		return &kebench.BenchMessage{Msg: ""}, nil
	}
	bf.Reset(int(length))
	_, err = io.ReadFull(c.rw, bf.Slice)
	if nil != err {
		return nil, err
	}
	tmp := make([]byte, length)
	copy(tmp, bf.Slice)
	return &kebench.BenchMessage{Msg: unsafe.String(&tmp[0], length)}, nil
}

func (c *EchoCodec3) Encode(msg *kebench.BenchMessage) error {
	bf := bufferPool.Get().(*Buffer)
	bf.Reset(4 + len(msg.Msg))
	defer func() {
		bufferPool.Put(bf)
	}()
	binary.BigEndian.PutUint32(bf.Slice, uint32(len(msg.Msg)))
	data := unsafe.Slice(unsafe.StringData(msg.Msg), len(msg.Msg))
	copy(bf.Slice[4:], data)
	_, err := c.rw.Write(bf.Slice)
	return err
}

var (
	allocPool [limit + 1]*sync.Pool
)

const (
	// Byte       = 1
	// KByte      = 1024 * Byte  // 1 << 10
	// MByte      = 1024 * KByte // 1 << 20
	// GByte      = 1024 * MByte // 1 << 30
	limit      = 26 // 1 << 26 64M
	limitBlock = 1 << limit
)

func init() {
	// 超过limit的大块数据，不通过池子获取
	allocPool[0] = &sync.Pool{
		New: func() any {
			return make([]byte, 0)
		},
	}
	for i := 1; i <= limit; i++ {
		block := 1 << (i - 1)
		allocPool[i] = &sync.Pool{
			New: func() any {
				return make([]byte, 0, block)
			},
		}
	}
}

func Malloc(size int, cap ...int) []byte {
	c := size
	if 0 != len(cap) && cap[0] > size {
		c = cap[0]
	}
	if c > limitBlock {
		return make([]byte, size, c)
	}
	idx := bits.Len32(uint32(c))
	if c&-c != c {
		idx += 1
	}
	get := allocPool[idx].Get().([]byte)
	return get[:size]
}

func Free(in []byte) {
	c := cap(in)
	if c > limitBlock {
		return
	}
	idx := bits.Len32(uint32(c))
	if c&-c != c {
		idx++
	}
	allocPool[idx].Put(in[:0])
}

type EchoCodec4 struct {
	rw io.ReadWriter
}

func NewEchoCodec4(rw io.ReadWriter) Codec {
	return &EchoCodec4{rw: rw}
}

func (c *EchoCodec4) Decode() (*kebench.BenchMessage, error) {
	head := Malloc(4)
	defer Free(head)
	_, err := io.ReadFull(c.rw, head[:])
	if nil != err {
		return nil, err
	}
	length := binary.BigEndian.Uint32(head[:])
	if length == 0 {
		return &kebench.BenchMessage{Msg: ""}, nil
	}
	data := Malloc(int(length))
	defer Free(data)
	_, err = io.ReadFull(c.rw, data)
	if nil != err {
		return nil, err
	}
	tmp := make([]byte, length)
	copy(tmp, data)
	msg := unsafe.String(&tmp[0], len(tmp))
	return &kebench.BenchMessage{Msg: msg}, nil
}

func (c *EchoCodec4) Encode(msg *kebench.BenchMessage) error {
	body := Malloc(4 + len(msg.Msg))
	defer Free(body)
	binary.BigEndian.PutUint32(body[:4], uint32(len(msg.Msg)))
	data := unsafe.Slice(unsafe.StringData(msg.Msg), len(msg.Msg))
	copy(body[4:], data)
	_, err := c.rw.Write(body)
	return err
}
