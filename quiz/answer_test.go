package quiz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func prepareCand() Candidate {
	c := Candidate{
		level: EASY,
		qns:   make(map[difficulty][]Question),
	}
	c.qns[EASY] = []Question{
		{Id: "Q1"}, {Id: "Q2"}, {Id: "Q3"}, {Id: "Q4"},
	}
	c.qns[MEDIUM] = []Question{
		{Id: "Q5"}, {Id: "Q6"}, {Id: "Q7"}, {Id: "Q8"}, {Id: "Q16"}, {Id: "Q17"},
	}
	c.qns[HARD] = []Question{
		{Id: "Q9"}, {Id: "Q10"}, {Id: "Q11"}, {Id: "Q12"}, {Id: "Q13"}, {Id: "Q14"}, {Id: "Q15"},
	}
	return c
}

func TestGoodCand(t *testing.T) {
	c := prepareCand()
	calibrateLevel(&c, true)
	require.Equal(t, 1, c.streak, "They should be equal.")

	calibrateLevel(&c, true)
	require.Equal(t, 2, c.streak, "They should be equal.")
	require.Equal(t, EASY, c.level, "They should be equal.")

	calibrateLevel(&c, true)
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, MEDIUM, c.level, "They should be equal.")

	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, HARD, c.level, "They should be equal.")

	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	require.Equal(t, 3, c.streak, "They should be equal.")
	require.Equal(t, HARD, c.level, "They should be equal.")

	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	require.Equal(t, 6, c.streak, "They should be equal.")
	require.Equal(t, HARD, c.level, "They should be equal.")

	calibrateLevel(&c, true)
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, EASY, c.level, "They should be equal.")

	calibrateLevel(&c, true)
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, MEDIUM, c.level, "They should be equal.")
}

func TestBumpyRide(t *testing.T) {
	c := prepareCand()
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)

	// moves to medium
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)

	// moves to hard
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, HARD, c.level, "They should be equal.") //
	calibrateLevel(&c, false)
	calibrateLevel(&c, false)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, false)
	calibrateLevel(&c, false)
	require.Equal(t, -2, c.streak, "They should be equal.")
	require.Equal(t, HARD, c.level, "They should be equal.")
	calibrateLevel(&c, false)

	// moves back to medium.
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, MEDIUM, c.level, "They should be equal.")
	calibrateLevel(&c, false)
	calibrateLevel(&c, false)
	calibrateLevel(&c, false)

	// moves to easy because negative streak == level streak.
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, EASY, c.level, "They should be equal.")

}

func TestBumpyRide2(t *testing.T) {
	c := prepareCand()
	// first question
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)
	calibrateLevel(&c, true)

	// moves to medium
	calibrateLevel(&c, false)
	require.Equal(t, -1, c.streak, "They should be equal.")
	require.Equal(t, MEDIUM, c.level, "They should be equal.")
	calibrateLevel(&c, true)
	calibrateLevel(&c, false)
	calibrateLevel(&c, true)
	calibrateLevel(&c, false)
	calibrateLevel(&c, true)

	// Medium questions finish, we move to hard ones now.
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, HARD, c.level, "They should be equal.")
	calibrateLevel(&c, true)
	calibrateLevel(&c, false)
	calibrateLevel(&c, true)
	calibrateLevel(&c, false)
	calibrateLevel(&c, true)
	calibrateLevel(&c, false)
	calibrateLevel(&c, true)

	// Hard finish, back to EASY.
	require.Equal(t, 0, c.streak, "They should be equal.")
	require.Equal(t, EASY, c.level, "They should be equal.")
}
