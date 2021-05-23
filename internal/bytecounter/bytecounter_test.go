package bytecounter

import (
	"bytes"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteCounter(t *testing.T) {
	r := bytes.NewBufferString("0123456789")
	w := bytes.NewBuffer(nil)

	tw := iotest.TruncateWriter(w, 4)
	bc := New(tw)

	_, err := io.Copy(bc, r)
	require.NoError(t, err)
	assert.Len(t, w.Bytes(), 4)
	assert.Equal(t, 4, bc.N)
}
