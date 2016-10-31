package quiz

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
)

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
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
		sr.Write(w, "", "Your quiz has already finished.",
			http.StatusBadRequest)
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
		m.Set(`<_uid_:` + userId + `> <quiz_start> "` + c.quizStart.Format(timeLayout) + `" .`)
		_, err := dgraph.SendMutation(m.String())
		if err != nil {
			sr.Write(w, "", err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(c.qns) == 0 {
		// No more questions to ask. Client ends quiz when question id is END.
		q := Question{
			Id:    "END",
			Score: x.Truncate(c.score),
		}

		// Lets store that the user successfully completed the test.
		m := new(dgraph.Mutation)
		m.Set(`<_uid_:` + userId + `> <complete> "true" .`)
		_, err := dgraph.SendMutation(m.String())
		if err != nil {
			sr.Write(w, "", err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(q)
		if err != nil {
			sr.Write(w, err.Error(), "", http.StatusInternalServerError)
			return
		}
		w.Write(b)
		go sendReport(userId)
		return
	}

	qn := c.qns[0]
	// Truncate score to two decimal places.
	qn.Score = x.Truncate(c.score)
	shuffleOptions(qn.Options)

	if c.lastQnUid != "" && c.lastQnUid == qn.Id {
		qn.Cid = c.lastQnCuid
		server.MarshalAndWrite(w, &qn)
		return
	}

	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + userId + `> <candidate.question> <_new_:qn> .`)
	m.Set(`<_new_:qn> <question.uid> <_uid_:` + qn.Id + `> .`)
	m.Set(`<_uid_:` + qn.Id + `> <question.candidate> <_uid_:` + userId + `> .`)
	m.Set(`<_new_:qn> <question.asked> "` + time.Now().Format("2006-01-02T15:04:05Z07:00") + `" .`)
	m.Set(`<_uid_:` + userId + `> <candidate.lastqnuid> "` + qn.Id + `" .`)
	res, err := dgraph.SendMutation(m.String())
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
	m = new(dgraph.Mutation)
	m.Set(`<_uid_:` + userId + `> <candidate.lastqncuid> "` + res.Uids["qn"] + `" .`)
	res, err = dgraph.SendMutation(m.String())
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}

	c.lastQnUid = qn.Id
	updateMap(userId, c)
	server.MarshalAndWrite(w, &qn)
}
