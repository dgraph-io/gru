package quiz

import "math/rand"

func shuffleQuestions(qns []Question) {
	for i := range qns {
		j := rand.Intn(i + 1)
		qns[i], qns[j] = qns[j], qns[i]
	}
}

func shuffleOptions(opts []Answer) {
	for i := range opts {
		j := rand.Intn(i + 1)
		opts[i], opts[j] = opts[j], opts[i]
	}
}

func qnsAsked(qns []qids) []string {
	var uids []string
	for _, qn := range qns {
		uids = append(uids, qn.QuestionUid[0].Id)
	}
	return uids
}

func isCorrectAnswer(selected []string, actual []string, pos, neg float64) float64 {
	if selected[0] == "skip" {
		return 0
	}
	// For multiple choice qnstions, we have partial scoring.
	if len(actual) == 1 {
		if selected[0] == actual[0] {
			return pos
		}
		return -neg
	}
	var score float64
	for _, aid := range selected {
		correct := false
		for _, caid := range actual {
			if caid == aid {
				correct = true
				break
			}
		}
		if correct {
			score += pos
		} else {
			score -= neg
		}
	}
	return score
}

func calcScore(qns []qids) float64 {
	score := 0.0
	for _, qn := range qns {
		score += qn.Score
	}
	return score
}
