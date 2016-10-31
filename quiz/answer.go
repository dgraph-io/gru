package quiz

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
)

type questionAnswer struct {
	Question []struct {
		Answered string `json:"question.answered"`
	} `json:"candidate.question"`
}

// Queries Dgraph and checks if the candidate has already answered the question.
func alreadyAnswered(cuid string) (int, error) {
	q := `{
                candidate.question(_uid_:` + cuid + `) {
                        question.answered
                }
        }`

	var ca questionAnswer
	if err := dgraph.QueryAndUnmarshal(q, &ca); err != nil {
		return http.StatusInternalServerError, err
	}

	if len(ca.Question) != 1 || ca.Question[0].Answered != "" {
		return http.StatusBadRequest, fmt.Errorf("You have already answered this question.")

	}
	return http.StatusOK, nil
}

func getScore(selected []string, actual []string, pos, neg float64) float64 {
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

type queInfo struct {
	Question []struct {
		Negative float64 `json:"negative,string"`
		Positive float64 `json:"positive,string"`
		Correct  []struct {
			Uid string `json:"_uid_"`
		} `json:"question.correct"`
	} `json:"question"`
}

func checkAnswer(qid string, ansIds []string) (float64, error) {
	q := `{
		question(_uid_: ` + qid + `) {
			question.correct {
				_uid_
			}
			positive
			negative
		}
	}`

	var qi queInfo
	if err := dgraph.QueryAndUnmarshal(q, &qi); err != nil {
		return 0, err
	}
	if len(qi.Question) != 1 {
		return 0, fmt.Errorf("There should be just one question returned")
	}

	qn := qi.Question[0]
	correct := []string{}
	for _, answer := range qn.Correct {
		correct = append(correct, answer.Uid)
	}

	score := getScore(ansIds, correct, qn.Positive, qn.Negative)
	return score, nil
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	userId, err := validateToken(r)
	if err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
		return
	}

	if status, err := checkAndUpdate(userId); err != nil {
		sr.Write(w, "", err.Error(), status)
		return
	}

	c, err := readMap(userId)
	if err != nil {
		sr.Write(w, "", "Candidate not found.", http.StatusBadRequest)
		return
	}

	if timeLeft(c.quizStart, c.quizDuration) < 0 {
		sr.Write(w, "", "Your quiz has already finished.", http.StatusBadRequest)
		return
	}

	qid := r.PostFormValue("qid")
	aid := r.PostFormValue("aid")
	cuid := r.PostFormValue("cuid")
	ansIds := strings.Split(aid, ",")
	if qid == "" || cuid == "" || len(ansIds) == 0 {
		sr.Write(w, "Answer ids/cuid/qid can't be empty", "", http.StatusBadRequest)
		return
	}

	// Since answering a question changes the score, we check if the candidate
	// has already answered this question.
	if status, err := alreadyAnswered(cuid); err != nil {
		sr.Write(w, err.Error(), "", status)
		return
	}

	// Lets get information about the question to check if the answer is right
	// and calculate scores.
	s, err := checkAnswer(qid, ansIds)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
	}
	c.score = c.score + s
	c.qns = c.qns[1:]
	updateMap(userId, c)

	// Lets store some information about this question.
	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + cuid + `> <candidate.answer> "` + aid + `" .`)
	m.Set(`<_uid_:` + cuid + `> <candidate.score> "` + strconv.FormatFloat(s, 'g', -1, 64) + `" .`)
	m.Set(`<_uid_:` + cuid + `> <question.answered> "` + time.Now().Format("2006-01-02T15:04:05Z07:00") + `" .`)
	if _, err = dgraph.SendMutation(m.String()); err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
}
