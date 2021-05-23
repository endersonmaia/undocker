package bytecounter

import "io"

// ByteCounter is an io.Writer that counts bytes written to it
type ByteCounter struct {
	N int64
	w io.Writer
}

// New returns a new ByteCounter
func New(w io.Writer) *ByteCounter {
	return &ByteCounter{w: w}
}

// Write writes to the underlying io.Writer and counts total written bytes
func (b *ByteCounter) Write(data []byte) (n int, err error) {
	defer func() { b.N += int64(n) }()
	return b.w.Write(data)
}
