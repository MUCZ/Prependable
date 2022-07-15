package prependable

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

// todo : 100% coverage
type MockReader struct{}

func (r *MockReader) Read(p []byte) (n int, err error) {
	if cap(p) < ReallyReadDataLength {
		panic(fmt.Sprint("wrong setting ", cap(p), "!=", ReallyReadDataLength))
	}
	copy(p, somedata[:ReallyReadDataLength])
	return ReallyReadDataLength, nil
}

var header []byte
var somedata []byte
var hl = 100
var readBufferLength = 4096
var ReallyReadDataLength int

func init() {
	ReallyReadDataLength, _ = strconv.Atoi(os.Getenv("RBL"))
	fmt.Println("readBufferLength:", ReallyReadDataLength)

	header = make([]byte, 0, hl)
	for i := 0; i < hl; i++ {
		header = append(header, byte(i))
	}
	somedata = make([]byte, 0, readBufferLength)
	for i := 0; i < hl; i++ {
		somedata = append(somedata, byte(i))
	}
}

func BenchmarkReadPrependable(b *testing.B) {
	r := MockReader{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lbf := New(hl, hl+readBufferLength)
		n, err := lbf.ReadFromReader(&r)
		if n != ReallyReadDataLength || err != nil {
			panic("read error")
		}
		lbf.Prepend(header)
	}
}

func BenchmarkRead(b *testing.B) {
	r := MockReader{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, readBufferLength, readBufferLength)
		n, _ := r.Read(buf)
		if n != ReallyReadDataLength {
			panic("read error")
		}
		buf = buf[:n]
		out := make([]byte, len(header)+len(buf))
		copy(out, []byte(header))
		copy(out[len(header):], buf)
	}
}

func TestBuf(t *testing.T) { // todo : use table
	buf := New(10, 100)
	if buf.begin != 10 || buf.end != 10 || buf.cap != 100 {
		t.Errorf("begin end cap:  expected %d %d %d ;what we get: %d %d %d", 10, 10, 100, buf.begin, buf.end, buf.cap)
	}

	tmp := []byte("hello")
	buf.Prepend(tmp)

	if buf.begin != 5 || buf.end != 10 || buf.cap != 100 {
		t.Errorf("begin end cap:  expected %d %d %d ;what we get: %d %d %d", 5, 10, 100, buf.begin, buf.end, buf.cap)
	}
	for i := range buf.View() {
		if len(buf.View()) != len(tmp) || buf.View()[i] != tmp[i] {
			t.Errorf("buf.View() wrong, expected %d, what we get %d", tmp, buf.View())
		}
	}
}
