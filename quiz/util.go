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

func qnsAnswered(qns []qids) []string {
	var uids []string
	for _, qn := range qns {
		if qn.Answered != "" {
			uids = append(uids, qn.QuestionUid[0].Id)
		}
	}
	return uids
}

func calcScore(qns []qids) float64 {
	score := 0.0
	for _, qn := range qns {
		score += qn.Score
	}
	return score
}
