package prependable

import (
	"testing"
)

// todo : 100% coverage

type MockReader struct {
	dataLen int
}

func (r *MockReader) Read(p []byte) (n int, err error) {
	n = r.dataLen
	if len(p) < r.dataLen {
		n = len(p)
	}
	copy(p, somedata[:n])
	return n, nil
}

var header []byte
var headerLen = 70
var readBufferLength = 4096

var somedata []byte

func init() {
	// make a header
	header = make([]byte, 0, headerLen)
	for i := 0; i < headerLen; i++ {
		header = append(header, byte(i))
	}
	// make some data for the mockReader
	somedata = make([]byte, 0, readBufferLength)
	for i := 0; i < readBufferLength; i++ {
		somedata = append(somedata, byte(i))
	}
}

func BenchmarkReadAndBuildPacket_Prependable(b *testing.B) {
	datalen := 2500
	r := MockReader{dataLen: datalen}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lbf := New(headerLen, headerLen+readBufferLength)
		n, err := lbf.ReadFromReader(&r)
		if n != datalen || err != nil {
			panic("read error")
		}
		lbf.Prepend(header)
	}
}

func BenchmarkReadAndBuildPacket_ByteSlice(b *testing.B) {
	datalen := 2500
	r := MockReader{dataLen: datalen}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, readBufferLength, readBufferLength)
		n, _ := r.Read(buf)
		if n != datalen {
			panic("read error")
		}
		buf = buf[:n]
		out := make([]byte, len(header)+len(buf))
		copy(out, []byte(header))
		copy(out[len(header):], buf)
	}
}

func BenchmarkPreTrim_Copy(b *testing.B) {
	datalen := 1500
	r := MockReader{dataLen: datalen}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, datalen)
		n, _ := r.Read(buf)
		if n != datalen {
			panic("read error")
		}
		copy(buf, buf[headerLen:datalen])
		// buf = buf[headerLen:datalen] // same efficency as above
	}
}

func BenchmarkPreTrim_Prependable(b *testing.B) {
	datalen := 1500
	r := MockReader{dataLen: datalen}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lbf := New(0, datalen)
		n, err := lbf.ReadFromReader(&r)
		if n != datalen || err != nil {
			panic("read error")
		}
		lbf.PreTrim(headerLen)
	}
}

func TestPrepend(t *testing.T) { // todo : use table tests
	bufs := make([]*Prependable, 0)
	// test both `New` and `NewFromSlice`
	bufs = append(bufs, New(10, 100))
	bufs = append(bufs, NewFromSlice(10, make([]byte, 100)))
	for _, buf := range bufs {
		if buf.readerIndex != 10 || buf.writerIndex != 10 || buf.size != 100 {
			t.Errorf("begin end cap:  expected %d %d %d ;what we get: %d %d %d", 10, 10, 100, buf.readerIndex, buf.writerIndex, buf.size)
		}

		// prepend 5 bytes
		tmp := []byte("hello")
		buf.Prepend(tmp)

		if buf.readerIndex != 5 || buf.writerIndex != 10 || buf.size != 100 {
			t.Errorf("begin end cap:  expected %d %d %d ;what we get: %d %d %d", 5, 10, 100, buf.readerIndex, buf.writerIndex, buf.size)
		}
		if len(buf.View()) != len(tmp) {
			t.Errorf("buf.View() wrong, expected %d, what we get %d", tmp, buf.View())
		}
		for i := range buf.View() {
			if buf.View()[i] != tmp[i] {
				t.Errorf("buf.View() wrong, expected %d, what we get %d", tmp, buf.View())
			}
		}

		// prepend 5 more bytes
		buf.Prepend(tmp)
		if buf.readerIndex != 0 || buf.writerIndex != 10 || buf.size != 100 {
			t.Errorf("begin end cap:  expected %d %d %d ;what we get: %d %d %d", 0, 10, 100, buf.readerIndex, buf.writerIndex, buf.size)
		}
		expected := []byte("hellohello")
		if len(buf.View()) != len(expected) {
			t.Errorf("buf.View() wrong, expected %d, what we get %d", tmp, buf.View())
		}
		for i := range buf.View() {
			if buf.View()[i] != expected[i] {
				t.Errorf("buf.View() wrong, expected %d, what we get %d", tmp, buf.View())
			}
		}

		// no more prependable space, error should be returned
		err := buf.Prepend(tmp)
		if err == nil {
			t.Errorf("Prepend should return error, but we get nil")
		}
	}
}

func TestGetterSetter(t *testing.T) {
	size := 100
	reserve := 10
	buf := New(reserve, size)
	if buf.Size() != size {
		t.Errorf("buf.Size() wrong, expected %d, what we get %d", size, buf.Size())
	}
	if buf.Prependable() != reserve {
		t.Errorf("buf.Prependable() wrong, expected %d, what we get %d", reserve, buf.Prependable())
	}
	if buf.Readable() != 0 {
		t.Errorf("buf.Readable() wrong, expected %d, what we get %d", 0, buf.Readable())
	}
	if buf.Writeable() != size-reserve {
		t.Errorf("buf.Writeable() wrong, expected %d, what we get %d", size-reserve, buf.Writeable())
	}
	if buf.Raw() == nil {
		t.Errorf("buf.Raw() wrong, expected not nil , what we get is nil")
	}
}

func TestReadAndPrepend(t *testing.T) {
	datalen := 1500
	r := MockReader{dataLen: datalen}
	buffer := New(headerLen, headerLen+readBufferLength)

	// read from reader
	n, err := buffer.ReadFromReader(&r)
	if n != datalen {
		t.Errorf("read error, n!=datalen, n=%d, datalen=%d", n, datalen)
	}
	if err != nil {
		t.Errorf("read error, err=%s", err)
	}

	// prepend a header
	err = buffer.Prepend(header)
	if err != nil {
		t.Errorf("prepend error, err=%s", err)
	}

	// check
	if buffer.Size() != (headerLen + readBufferLength) {
		t.Errorf("buffer.Size() wrong, expected %d, what we get %d", headerLen+readBufferLength, buffer.Size())
	}
	if buffer.Readable() != datalen+headerLen {
		t.Errorf("buffer.Readable() wrong, expected %d, what we get %d", datalen+headerLen, buffer.Readable())
	}
	if buffer.Prependable() != 0 {
		t.Errorf("buffer.Prependable() wrong, expected %d, what we get %d", 0, buffer.Prependable())
	}
	if buffer.Writeable() != readBufferLength-datalen {
		t.Errorf("buffer.Writeable() wrong, expected %d, what we get %d", readBufferLength-datalen, buffer.Writeable())
	}

	// compare result(what we get) and expected byte by byte
	expected := make([]byte, headerLen+datalen)
	copy(expected[:headerLen], header)
	r.Read(expected[headerLen:])

	get := buffer.View()
	if len(get) != len(expected) {
		t.Errorf("buffer.View() len wrong, expected %d, what we get %d", len(expected), len(get))
	}
	for i := range expected {
		if get[i] != expected[i] {
			t.Errorf("buffer.View() wrong, expected %d, \nwhat we get %d", expected, buffer.View())
			break
		}
	}
}

// read first, then prepend a header
func TestReadN(t *testing.T) {
	datalen := 1500
	r := MockReader{dataLen: datalen}
	buffer := New(headerLen, headerLen+datalen)

	// read N from reader
	N := 100
	n, err := buffer.ReadNbytesFromReader(&r, N)
	if n != N {
		t.Errorf("read error, n!=datalen, n=%d, datalen=%d", n, N)
	}
	if err != nil {
		t.Errorf("read error, err=%s", err)
	}

	// prepend a header
	err = buffer.Prepend(header)
	if err != nil {
		t.Errorf("prepend error, err=%s", err)
	}

	// check sizes
	if buffer.Size() != (headerLen + datalen) {
		t.Errorf("buffer.Size() wrong, expected %d, what we get %d", headerLen+readBufferLength, buffer.Size())
	}
	if buffer.Readable() != headerLen+N {
		t.Errorf("buffer.Readable() wrong, expected %d, what we get %d", datalen+headerLen, buffer.Readable())
	}
	if buffer.Prependable() != 0 {
		t.Errorf("buffer.Prependable() wrong, expected %d, what we get %d", 0, buffer.Prependable())
	}
	if buffer.Writeable() != datalen-N {
		t.Errorf("buffer.Writeable() wrong, expected %d, what we get %d", readBufferLength-datalen, buffer.Writeable())
	}

	// compare result(what we get) and expected byte by byte
	expected := make([]byte, headerLen+N)
	copy(expected[:headerLen], header)
	r.Read(expected[headerLen:])

	get := buffer.View()
	if len(get) != len(expected) {
		t.Errorf("buffer.View() len wrong, expected %d, what we get %d", len(expected), len(get))
	}
	for i := range expected {
		if get[i] != expected[i] {
			t.Errorf("buffer.View() wrong, expected %d, \nwhat we get %d", expected, buffer.View())
			break
		}
	}

	// test over read
	N = datalen - N + 1
	n, err = buffer.ReadNbytesFromReader(&r, N)
	if err == nil {
		t.Errorf("ReadNBytesFromReader should return error, but we get nil")
	}
}

// prepend header first, then read
func TestReadNReverse(t *testing.T) {
	datalen := 1500
	r := MockReader{dataLen: datalen}
	buffer := New(headerLen, headerLen+readBufferLength)

	// prepend a header first
	err := buffer.Prepend(header)
	if err != nil {
		t.Errorf("prepend error, err=%s", err)
	}

	// read N from reader then
	N := 100
	n, err := buffer.ReadNbytesFromReader(&r, N)
	if n != N {
		t.Errorf("read error, n!=datalen, n=%d, datalen=%d", n, N)
	}
	if err != nil {
		t.Errorf("read error, err=%s", err)
	}

	// check sizes
	if buffer.Size() != (headerLen + readBufferLength) {
		t.Errorf("buffer.Size() wrong, expected %d, what we get %d", headerLen+readBufferLength, buffer.Size())
	}
	if buffer.Readable() != headerLen+N {
		t.Errorf("buffer.Readable() wrong, expected %d, what we get %d", datalen+headerLen, buffer.Readable())
	}
	if buffer.Prependable() != 0 {
		t.Errorf("buffer.Prependable() wrong, expected %d, what we get %d", 0, buffer.Prependable())
	}
	if buffer.Writeable() != readBufferLength-N {
		t.Errorf("buffer.Writeable() wrong, expected %d, what we get %d", readBufferLength-datalen, buffer.Writeable())
	}

	// compare result(what we get) and expected byte by byte
	expected := make([]byte, headerLen+N)
	copy(expected[:headerLen], header)
	r.Read(expected[headerLen:])

	get := buffer.View()
	if len(get) != len(expected) {
		t.Errorf("buffer.View() len wrong, expected %d, what we get %d", len(expected), len(get))
	}
	for i := range expected {
		if get[i] != expected[i] {
			t.Errorf("buffer.View() wrong, expected %d, \nwhat we get %d", expected, buffer.View())
			break
		}
	}
}

func TestFull(t *testing.T) {
	datalen := 1500
	r := MockReader{dataLen: datalen}
	buffer := New(headerLen, headerLen+datalen)
	// read datalen from reader
	n, err := buffer.ReadFromReader(&r)
	if n != datalen {
		t.Errorf("read error, n!=datalen, n=%d, datalen=%d", n, datalen)
	}
	if err != nil {
		t.Errorf("read error, err=%s", err)
	}
	// prepend a header
	err = buffer.Prepend(header)
	if err != nil {
		t.Errorf("prepend error, err=%s", err)
	}

	// should have no more space
	if buffer.Writeable() != 0 {
		t.Errorf("buffer.Writeable() wrong, expected %d, what we get %d", 0, buffer.Writeable())
	}
	if buffer.Prependable() != 0 {
		t.Errorf("buffer.Prependable() wrong, expected %d, what we get %d", 0, buffer.Prependable())
	}
	// should have error
	n, err = buffer.ReadFromReader(&r)
	if err == nil {
		t.Errorf("read error, err=nil,expected to be not nil because the buffer should be full")
	}
	n, err = buffer.ReadNbytesFromReader(&r, 1)
	if err == nil {
		t.Errorf("read error, err=nil,expected to be not nil because the buffer should be full")
	}
}

func TestMultiTimesPrepen(t *testing.T) {
	buffer := New(headerLen, headerLen+1)
	var err error

	for i := 0; i < headerLen; i++ {
		err = buffer.Prepend([]byte{byte(i)})
		if err != nil {
			t.Errorf("prepend error, err=%s", err)
		}
	}

	// no more prependable space, should return errror
	err = buffer.Prepend([]byte{byte(0)})
	if err == nil {
		t.Errorf("prepend error, err=nil,expected to be not nil because the buffer should be full")
	}
}

func TestPreTrim(t *testing.T) {
	data := []byte("hello")
	buffer := New(headerLen, headerLen+1024)
	// prepend a header
	buffer.Prepend(data)
	v := buffer.View()
	// compare
	if len(v) != len(data) {
		t.Errorf("buffer.View() wrong, expected %d, what we get %d", len(data), len(v))
	}
	for i := range v {
		if data[i] != v[i] {
			t.Errorf("buffer.View() wrong, expected %d, \nwhat we get %d", data, buffer.View())
			break
		}
	}

	// test pretrim
	for j := 0; j <= len(data); j++ {
		// pretrim one byte = one letter each time
		data = data[1:]
		buffer.PreTrim(1)
		v = buffer.View()
		if len(v) != len(data) {
			t.Errorf("buffer.View() wrong, expected %d, what we get %d", len(data), len(v))
		}
		for i := range v {
			if data[i] != v[i] {
				t.Errorf("buffer.View() wrong, expected %d, \nwhat we get %d", data, buffer.View())
				break
			}
		}
	}
}

// for the read operation, Prependable buffer should be similarly efficient to []byte

func BenchmarkByteSlice_ControlGroup(b *testing.B) {
	b.ResetTimer()
	r := MockReader{readBufferLength - headerLen}
	for i := 0; i < b.N; i++ {
		// prepare some data
		payload := make([]byte, readBufferLength)
		n, err := r.Read(payload)
		if n != readBufferLength-headerLen || err != nil {
			panic(err)
		}
	}
}

func BenchmarkNewFromSlice_ControlGroup(b *testing.B) {
	b.ResetTimer()
	r := MockReader{readBufferLength - headerLen}
	for i := 0; i < b.N; i++ {
		// prepare some data
		payload := make([]byte, readBufferLength)
		pp := NewFromSlice(headerLen, payload)
		n, err := pp.ReadFromReader(&r)
		if n != readBufferLength-headerLen || err != nil {
			panic(err)
		}
	}
}

func BenchmarkNew_ControlGroup(b *testing.B) {
	b.ResetTimer()
	r := MockReader{readBufferLength - headerLen}
	for i := 0; i < b.N; i++ {
		// prepare some data
		pp := New(headerLen, readBufferLength)
		n, err := pp.ReadFromReader(&r)
		if n != readBufferLength-headerLen || err != nil {
			panic(err)
		}
	}
}
