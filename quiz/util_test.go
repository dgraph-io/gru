package quiz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsCorrectAnswer(t *testing.T) {
	score := getScore([]string{"ringplanets-saturn", "ringplanets-neptune"}, []string{"ringplanets-saturn"}, 2.5, 3)
	require.Equal(t, 2.5, score)

	score = getScore([]string{"ringplanets-saturn", "ringplanets-neptune"}, []string{"ringplanets-saturn", "ringplanets-neptune"}, 2.5, 3)
	require.Equal(t, 5.0, score)

	score = getScore([]string{"ringplanets-venus", "ringplanets-mars"}, []string{"ringplanets-saturn", "ringplanets-neptune"}, 2.5, 3)
	require.Equal(t, -6.0, score)

	score = getScore([]string{"skip"}, []string{"ringplanets-saturn", "ringplanets-neptune"}, 2.5, 3)
	require.Equal(t, 0.0, score)

	// Single choice questions
	score = getScore([]string{"sunmass-99"}, []string{"sunmass-99"}, 5.0, 2.5)
	require.Equal(t, 5.0, score)

	score = getScore([]string{"sunmass-1"}, []string{"sunmass-99"}, 5.0, 2.5)
	require.Equal(t, -2.5, score)
}
