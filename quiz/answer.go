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
        candidate.question(id:` + cuid + `) {
            question.answered
        }
    }`

	var ca questionAnswer
	if err := dgraph.QueryAndUnmarshal(q, &ca); err != nil {
		return http.StatusInternalServerError, err
	}

	if len(ca.Question) > 0 && ca.Question[0].Answered != "" {
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
		Negative float64 `json:"negative"`
		Positive float64 `json:"positive"`
		Correct  []struct {
			Uid string `json:"_uid_"`
		} `json:"question.correct"`
	} `json:"question"`
}

func checkAnswer(qid string, ansIds []string) (float64, bool, error) {
	q := `{
		question(id: ` + qid + `) {
			question.correct {
				_uid_
			}
			positive
			negative
		}
	}`

	var qi queInfo
	if err := dgraph.QueryAndUnmarshal(q, &qi); err != nil {
		return 0, false, err
	}
	if len(qi.Question) != 1 {
		return 0, false, fmt.Errorf("There should be just one question returned")
	}

	qn := qi.Question[0]
	correctAids := []string{}
	for _, answer := range qn.Correct {
		correctAids = append(correctAids, answer.Uid)
	}

	score := getScore(ansIds, correctAids, qn.Positive, qn.Negative)
	correct := (score == float64(len(correctAids))*qn.Positive)
	return score, correct, nil
}

const (
	// Positive or negative streak at which we change the level for the candidate.
	levelStreak = 3
	// EASY, MEDIUM and HARD.
	numLevels = 3
)

func (c *Candidate) updateStreak(correct bool) {
	if correct {
		if c.streak < 0 {
			c.streak = 1
		} else {
			c.streak++
		}
	} else {
		if c.streak > 0 {
			c.streak = -1
		} else {
			c.streak--
		}
	}
}

func (c *Candidate) downgradeLevel() {
	currentLevel := c.level
	for i := 1; i < numLevels; i++ {
		newLevel := currentLevel - difficulty(i)
		if newLevel < 0 {
			newLevel += numLevels
		}
		if len(c.qns[newLevel]) != 0 {
			c.level = newLevel
			// We are downgrading the level, lets reset the streak.
			c.streak = 0
			break
		}

	}
}

func (c *Candidate) upgradeLevel() {
	// Lets check if we have questions available for the next level.
	for i := 1; i < numLevels; i++ {
		newLevel := difficulty((int(c.level) + i) % numLevels)
		if len(c.qns[newLevel]) != 0 {
			c.level = newLevel
			// We are upgrading the level, lets reset the streak.
			c.streak = 0
			break
		}
	}
}

func calibrateLevel(c *Candidate, correct bool) {
	c.updateStreak(correct)

	// Lets delete the first question for this level, since the candidate
	// has already answered it.
	c.qns[c.level] = c.qns[c.level][1:]
	// If user has a negative streak and his current level is not EASY, we
	// downgrade his level.
	if c.streak == -levelStreak && c.level != EASY {
		c.downgradeLevel()
		// If user has a positive streak or if we run out of questions for this
		// level, we try to upgrade the level.
	} else if (c.streak == levelStreak && c.level != HARD) || len(c.qns[c.level]) == 0 {
		c.upgradeLevel()
	}
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	userId, err := validateToken(r)
	if err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
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
	s, correct, err := checkAnswer(qid, ansIds)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
	}
	c.score = c.score + s
	calibrateLevel(&c, correct)
	c.lastQnAsked = time.Now().UTC()
	updateMap(userId, c)

	// Lets store some information about this question.
	m := new(dgraph.Mutation)
	m.Set(`<` + cuid + `> <candidate.answer> "` + aid + `" .`)
	m.Set(`<` + cuid + `> <candidate.score> "` + strconv.FormatFloat(s, 'g', -1, 64) + `" .`)
	m.Set(`<` + cuid + `> <question.answered> "` + time.Now().UTC().Format(timeLayout) + `" .`)

	if _, err = dgraph.SendMutation(m.String()); err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
