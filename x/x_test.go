package x

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncate(t *testing.T) {
	var f float64
	f = 4.62
	f = Truncate(4.6)
	require.Equal(t, 4.60, f, "Values should be equal.")
}
