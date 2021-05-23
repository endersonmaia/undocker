package bytecounter

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteCounter(t *testing.T) {
	r := bytes.NewBufferString("0123456789")
	w := bytes.NewBuffer(nil)
	bc := New(w)
	_, err := io.Copy(bc, r)
	require.NoError(t, err)
	assert.Equal(t, []byte("0123456789"), w.Bytes())
	assert.Equal(t, int64(10), bc.N)
}

func TestSingleByteCounter(t *testing.T) {
	r := bytes.NewBufferString("0123456789")
	w := bytes.NewBuffer(nil)
	tw := &shortWriter{w}
	bc := New(tw)
	_, err := io.Copy(bc, r)
	assert.EqualError(t, err, "short write")
	assert.Len(t, w.Bytes(), 1)
	assert.Equal(t, int64(1), bc.N)
}

// shortWriter writes only first byte
type shortWriter struct {
	w io.Writer
}

// Write writes a byte to the underlying writer
func (f *shortWriter) Write(p []byte) (int, error) {
	return f.w.Write(p[0:1])
}
