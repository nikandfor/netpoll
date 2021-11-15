package netpoll

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	s, err := Select([]io.Reader{r}, 0)
	assert.Equal(t, ErrNoEvents, err)

	s, err = Select([]io.Reader{r}, time.Millisecond)
	assert.Equal(t, ErrNoEvents, err)

	go func() {
		_, err := w.Write([]byte("data"))
		require.NoError(t, err)
	}()

	s, err = Select([]io.Reader{r}, -1)
	assert.NoError(t, err)
	assert.True(t, s == r, "bad file returned")
}
