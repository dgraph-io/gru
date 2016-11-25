package quiz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	qns := []Question{
		Question{
			Tags: []Tag{
				Tag{Name: "easy"},
				Tag{Name: "medium"},
			},
		},
		Question{
			Tags: []Tag{
				Tag{Name: "hard"},
				Tag{Name: "algorithms"},
			},
		},
		Question{
			Tags: []Tag{
				Tag{Name: "systems"},
				Tag{Name: "easy"},
			},
		},
		Question{
			Tags: []Tag{
				Tag{Name: "medium"},
				Tag{Name: "timecomplexity"},
			},
		},
	}

	qnMap := filter(qns)

	require.Equal(t, 2, len(qnMap[EASY]), "We should have 2 easy questions.")
	require.Equal(t, 1, len(qnMap[MEDIUM]), "We should have 1 medium question.")
	require.Equal(t, 1, len(qnMap[HARD]), "We should have 1 hard question.")
}
