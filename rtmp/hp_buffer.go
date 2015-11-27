// The MIT License (MIT)
//
// Copyright (c) 2014 winlin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package rtmp

// to cache bytes
// user can use the bytes buffer like a list:
// list.Append([1, 2, 3]), the list.Bytes() is [1, 2, 3]
// list.Remove(2), the list.Bytes() is [3]
// list.Append([1, 2]), the list.Bytes() is [3, 1, 2]
type BytesList struct {
	buf []byte
	start int
	end int
}
func NewBytesList(b []byte) (*BytesList) {
	r := &BytesList{}
	r.buf = b
	r.end = len(b)
	return r
}
// get the length of buffer, like the length of list.
func (r *BytesList) Len() (n int) {
	return r.end - r.start
}
// the bytes of buffer, like the bytes of list.
func (r *BytesList) Bytes() []byte {
	return r.buf[r.start:r.end]
}
// append bytes to the end of bytes.
func (r *BytesList) Append(b []byte) {
	// append bytes to the end of logic buffer
	exists_len := r.Len()
	r.grow_to(exists_len + len(b))

	exists_bytes := r.Bytes()
	copy(exists_bytes[exists_len:], b)
}
// remove n bytes, from the start of buf
// if all bytes removed, reset the start and end to zero
func (r *BytesList) Remove(n int) {
	if n <= 0 {
		return
	}

	if n >= r.Len() {
		r.start = 0
		r.end = 0
	} else {
		r.start += n
	}
}
// grow the end of bytes, ensure can use copy always,
// that is, ensure the Len() always greater than or equals to n
// append bytes to the end if need more space
func (r *BytesList) grow_to(n int) {
	if n <= 0 {
		return
	}

	// grow the capacity
	capacity_grow := n - (len(r.buf) - r.start)
	if capacity_grow > 0 {
		r.buf = append(r.buf, make([]byte, capacity_grow)...)
	}

	// grow the r.end to grow the Bytes()
	r.end += n - r.Len()
}

/**
* high performance bytes buffer, read and write from zero.
 */
type HPBuffer struct {
	buffer *BytesList
	off int
}
func NewHPBuffer(b []byte) (*HPBuffer) {
	r := &HPBuffer{}
	r.buffer = NewBytesList(b)
	return r
}
func (r *HPBuffer) String() string {
	if r == nil {
		return "<nil>"
	}
	return string(r.Bytes())
}
func (r *HPBuffer) Reset() {
	r.off = 0
}
func (r *HPBuffer) Len() (int) {
	return r.buffer.Len() - r.off
}
func (r *HPBuffer) Bytes() []byte {
	b := r.buffer.Bytes()
	return b[r.off:]
}
func (r *HPBuffer) WrittenBytes() ([]byte) {
	b := r.buffer.Bytes()
	return b[0:r.off]
}
func (r *HPBuffer) Append(b []byte) (n int, err error) {
	r.buffer.Append(b)

	// TODO: FIXME: return err
	return
}
func (r *HPBuffer) Consume(n int) (err error) {
	r.buffer.Remove(n)
	r.off -= n
	// TODO: FIXME: return err
	return
}
func (r *HPBuffer) Skip(n int) (err error) {
	r.off += n
	// TODO: FIXME: return err
	return
}
func (r *HPBuffer) Read(b []byte) (n int, err error) {
	bytes := r.Bytes()

	n = len(b)
	copy(b, bytes[0:n])
	r.off += n
	// TODO: FIXME: return err
	return
}
func (r *HPBuffer) Write(b []byte) (n int, err error) {
	bytes := r.Bytes()

	n = len(b)
	copy(bytes[0:n], b)
	r.off += n
	// TODO: FIXME: return err
	return
}
