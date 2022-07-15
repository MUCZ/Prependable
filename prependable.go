package prependable

import (
	"fmt"
	"io"
	"net"
)

type Prependable struct {
	v     *[]byte
	begin int // maybe uint or uint16
	end   int
	cap   int
}

func NewFromSlice(reserve int, v *[]byte) *Prependable {
	if reserve >= cap(*v)-1 {
		panic("before must be less than len(v)-1")
	}
	return &Prependable{v: v, begin: reserve, end: reserve, cap: cap(*v) - 1}
}

func New(reserve int, size int) *Prependable {
	if reserve >= size-1 {
		panic(fmt.Sprint("before must be less than size-1", reserve, size))
	}
	v := make([]byte, 0, size)
	return &Prependable{v: &v, begin: reserve, end: reserve, cap: size}
}

func (p *Prependable) GetRaw() *[]byte {
	return p.v
}

func (p *Prependable) readView() []byte {
	return (*p.v)[p.begin:p.end]
}

func (p *Prependable) View() []byte {
	return (*p.v)[p.begin:p.end]
}

func (p *Prependable) PrependView(size int) []byte {
	if size > p.begin {
		panic("size must be less than begin")
		// 扩容
	} else {
		p.begin -= size
		return (*p.v)[p.begin : p.begin+size] // 不能用超了
	}
}

func (p *Prependable) ReadFromConn(c net.Conn) (n int, err error) {
	return p.ReadFromReader(io.Reader(c))
}

func (p *Prependable) ReadFromReader(r io.Reader) (n int, err error) {
	if p.end == p.cap {
		return 0, fmt.Errorf("buffer is full")
	}
	n, err = r.Read(p.readView())
	if err != nil {
		return 0, err
	}
	if n != 0 {
		p.end += n
	}
	return n, nil
}

func (p *Prependable) Prepend(data []byte) {
	size := len(data)
	prespace := p.PrependView(size)
	copy(prespace, data)
}
