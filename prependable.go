package prependable

import (
	"fmt"
	"io"
	"net"
)

// Prependable is a buffer that support 'Prepend(data []byte)'
// method without copying and moving the existing payload inside
// the buffer.

// It is useful when building networking packets, where each
// protocol adds its own headers to the front of the
// higher-level protocol header and payload; for example, TCP
// would prepend its header to the payload, then IP would
// prepend its own, then ethernet.

// The larger the (len(payload)/len(header)), the more Prependable
// Buffer can show performance benefits.

var (
	ErrNoEnoughPrependSpace = fmt.Errorf("no enough space to prepend")
	ErrNoEnoughWriteSpace   = fmt.Errorf("no enough space to write")
	ErrFullBuffer           = fmt.Errorf("buffer is full")
	ErrCopyError            = fmt.Errorf("copy error: size not equal to expected")
	ErrReadNFailed          = fmt.Errorf("read bytes not equal to expected")
	ErrInvalidReserveBytes  = fmt.Errorf("reserves bytes greater than size or less than 0")
)

/// |- reserved  space -|
/// +-------------------+------------------+------------------+
/// | prependable bytes |  readable bytes  |  writable bytes  |
/// |                   |     (CONTENT)    |                  |
/// +-------------------+------------------+------------------+
/// |                   |                  |                  |
/// 0      <=      readerIndex   <=   writerIndex    <=     size

// todo: use reflect.SliceHeader to optimize this struct
type Prependable struct {
	v           []byte // buffer underlying
	readerIndex int
	writerIndex int
	size        int
}

func NewFromSlice(reserve int, v []byte) *Prependable {
	if reserve > cap(v) || reserve < 0 {
		panic(ErrInvalidReserveBytes)
	}
	return &Prependable{v: v, readerIndex: reserve, writerIndex: reserve, size: cap(v)}
}

func New(reserve int, size int) *Prependable {
	if reserve > size || reserve < 0 {
		panic(ErrInvalidReserveBytes)
	}
	v := make([]byte, 0, size)
	return &Prependable{v: v, readerIndex: reserve, writerIndex: reserve, size: size}
}

// get the underlying slice
func (p *Prependable) Raw() []byte { return p.v }

// getters
func (p *Prependable) Size() (size int)                   { return p.size }
func (p *Prependable) Readable() (readableSize int)       { return p.writerIndex - p.readerIndex }
func (p *Prependable) Writeable() (writableSize int)      { return p.size - p.writerIndex }
func (p *Prependable) Prependable() (prependableSize int) { return p.readerIndex }

// trim the content from left
func (p *Prependable) PreTrim(size int) {
	if size > p.Readable() {
		// trim All
		p.readerIndex = p.writerIndex
		return
	}
	p.readerIndex += size
}

func (p *Prependable) prependView(size int) []byte {
	return (p.v)[p.readerIndex-size : p.readerIndex]
}

// to get the content
func (p *Prependable) View() []byte {
	return (p.v)[p.readerIndex:p.writerIndex]
}
func (p *Prependable) writeView() []byte {
	return (p.v)[p.writerIndex:p.size]
}
func (p *Prependable) writeViewN(n int) []byte {
	return (p.v)[p.writerIndex : p.writerIndex+n]
}

func (p *Prependable) ReadFromConn(c net.Conn) (n int, err error) {
	return p.ReadFromReader(io.Reader(c))
}

func (p *Prependable) ReadFromReader(r io.Reader) (n int, err error) {
	if p.writerIndex == p.size {
		return 0, ErrFullBuffer
	}
	n, err = r.Read(p.writeView())
	if err != nil {
		return n, err
	}
	if n != 0 {
		p.writerIndex += n
	}
	return n, nil
}

// read n bytes from reader
func (p *Prependable) ReadNbytesFromReader(r io.Reader, readSize int) (n int, err error) {
	if readSize <= 0 {
		return 0, nil
	}
	if p.writerIndex == p.size {
		return 0, ErrFullBuffer
	}
	if readSize+p.writerIndex >= p.size {
		return 0, ErrNoEnoughWriteSpace
	}
	v := p.writeViewN(readSize)
	n, err = io.ReadFull(r, v)
	if err != nil {
		return n, err
	}
	p.writerIndex += n
	if n != readSize {
		return n, ErrReadNFailed
	}
	return n, nil
}

func (p *Prependable) Prepend(data []byte) error {
	size := len(data)
	if p.readerIndex < size {
		return ErrNoEnoughPrependSpace
	}
	prespace := p.prependView(size)
	n := copy(prespace, data)
	p.readerIndex -= size
	if n != size {
		return ErrCopyError
	}
	return nil
}
