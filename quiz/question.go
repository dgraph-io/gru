package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
)

type Tag struct {
	Name string `json:"name"`
}

// Question is marshalled to JSON and sent to the client.
type Question struct {
	Uid string `json:"uid"`

	// TODO - I wanted to write a sarcastic comment about this name choice,
	// but all my ideas were totally passive agressive
	// Rename cuid to something that makes sense without reading the below comment

	// cuid represents the uid of the question asked to the candidate, it is linked
	// to the original question uid.
	Cid     string     `json:"cuid"`
	Text    string		 `json:"text"`
	Options []Answer	 `json:"options"`
	IsMultiple bool    `json:"multiple"`
	Positive   float64 `json:"positive"`
	Negative   float64 `json:"negative"`
	// Score of the candidate is sent as part of the questions API.
	Score     float64 `json:"score"`
	TimeTaken string  `json:"time_taken"`
	// Current question number.
	Idx int `json:"idx"`
	// Total number of questions.
	NumQns int `json:"num_qns"`
}

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
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

	if c.quizStart.IsZero() {
		// This means its the first question he is being asked. Lets
		// store quizStart so that we can use to calculate timeLeft for
		// Ping API. Lets also persist it to database, so that we can
		// recover it incase we crash.
		c.quizStart = time.Now().UTC()
		updateMap(userId, c)
		m := new(dgraph.Mutation)
		m.SetString(userId, "quiz_start", c.quizStart.Format(time.RFC3339Nano))
		_, err := dgraph.SendMutation(m)
		if err != nil {
			sr.Write(w, "", err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(c.qns[c.level]) == 0 || c.score <= c.quizThreshold {
		// No more questions to ask or score is less than threshold.
		// Client ends quiz when question id is END.
		q := Question{
			Uid:    "END",
			Score: c.score,
		}

		// Lets store that the user successfully completed the test.
		m := new(dgraph.Mutation)
		m.SetString(userId, "complete", "true")
		// Completed at is used to reject candidates whose score is < cutoff
		m.SetString(userId, "completed_at", time.Now().Format(time.RFC3339Nano))
		m.SetString(userId, "score", strconv.FormatFloat(c.score, 'f', 2, 64))
		_, err := dgraph.SendMutation(m)
		if err != nil {
			sr.Write(w, "", err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(q)
		if err != nil {
			sr.Write(w, err.Error(), "", http.StatusInternalServerError)
			return
		}
		if err = sendMail(c, userId); err != nil {
			sr.Write(w, err.Error(), "", http.StatusInternalServerError)
			return
		}
		w.Write(b)
		return
	}

	qn := c.qns[c.level][0]
	if c.lastQnAsked.IsZero() {
		qn.TimeTaken = "0s"
		c.lastQnAsked = time.Now().UTC()
		updateMap(userId, c)
	} else {
		qn.TimeTaken = time.Now().UTC().Sub(c.lastQnAsked).String()
	}

	qn.Score = c.score
	shuffleOptions(qn.Options)

	qn.NumQns = c.numQuestions
	qn.Idx = c.qnIdx
	updateMap(userId, c)
	if c.lastQnUid != "" && c.lastQnUid == qn.Uid {
		qn.Cid = c.lastQnCuid
		server.MarshalAndWrite(w, &qn)
		return
	}

	m := new(dgraph.Mutation)
	m.SetLink(userId, "candidate.question", "_:qn")
	m.SetLink("_:qn", "question", qn.Uid)
	m.SetLink(qn.Uid, "question.candidate", userId)
	m.SetString("_:qn", "question.asked", time.Now().Format(time.RFC3339Nano))
	m.SetString(userId, "candidate.lastqnuid", qn.Uid)
	res, err := dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if res.Uids["qn"] == "" {
		sr.Write(w, res.Message, "", http.StatusInternalServerError)
		return
	}

	c.lastQnCuid = res.Uids["qn"]
	qn.Cid = res.Uids["qn"]
	c.lastQnUid = qn.Uid
	qn.Idx = c.qnIdx + 1
	c.qnIdx += 1
	updateMap(userId, c)
	server.MarshalAndWrite(w, &qn)
}
