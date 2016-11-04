package candidate

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/gorilla/mux"
)

type Candidate struct {
	Uid       string
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	Validity  string `json:"validity"`
	QuizId    string `json:"quiz_id"`
	OldQuizId string `json:"old_quiz_id"`
}

const (
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	validityLayout = "2006-01-02"
)

// TODO - Optimize later.
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func index(quizId string) string {
	return `
	{
	quiz(_uid_: ` + quizId + `) {
		quiz.candidate {
			_uid_
			name
			email
			validity
			complete
			quiz_start
			invite_sent
			candidate.question {
				candidate.score
			}
		}
	}
}
`
}

func Index(w http.ResponseWriter, r *http.Request) {
	quizId := r.URL.Query().Get("quiz_id")
	sr := server.Response{}
	if quizId == "" {
		sr.Write(w, "", "Quiz id can't be empty.", http.StatusBadRequest)
		return
	}
	q := index(quizId)
	res, err := dgraph.Query(q)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func add(c Candidate) string {
	return `
	mutation {
		set {
		<_uid_:` + c.QuizId + `> <quiz.candidate> <_new_:c> .
		<_new_:c> <candidate.quiz> <_uid_:` + c.QuizId + `> .
		<_new_:c> <email> "` + c.Email + `" .
		<_new_:c> <token> "` + c.Token + `" .
		<_new_:c> <validity> "` + c.Validity + `" .
		<_new_:c> <invite_sent> "` + time.Now().UTC().String() + `" .
		<_new_:c> <complete> "false" .
		}
	}`
}

func Add(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	var c Candidate
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var t time.Time
	if t, err = time.Parse(validityLayout, c.Validity); err != nil {
		sr.Message = "Couldn't parse the validity"
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	c.Validity = t.String()
	// TODO - Validate candidate fields shouldn't be empty.
	c.Token = randStringBytes(33)
	m := add(c)
	mr, err := dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}

	// mutation applied successfully, lets send a mail to the candidate.
	uid, ok := mr.Uids["c"]
	if !ok {
		sr.Write(w, "Uid not returned for newly created candidate by Dgraph.",
			"", http.StatusInternalServerError)
		return
	}

	// Token sent in mail is uid + the random string.
	go mail.Send(c.Name, c.Email, t.Format("Mon Jan 2 15:04:05 MST 2006"),
		uid+c.Token)
	sr.Message = "Candidate added successfully."
	sr.Success = true
	w.Write(server.MarshalResponse(sr))
}

func edit(c Candidate) string {
	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + c.Uid + `> <email> "` + c.Email + `" . `)
	m.Set(`<_uid_:` + c.Uid + `> <validity> "` + c.Validity + `" . `)

	// When the quiz for which candidate is invited is changed, we get both OldQuizId
	// and new QuizId.
	if c.QuizId != "" {
		m.Set(`<_uid_:` + c.QuizId + `> <quiz.candidate> <_uid_:` + c.Uid + `> .`)
		m.Set(`<_uid_:` + c.Uid + `> <candidate.quiz> <_uid_:` + c.QuizId + `> .`)
	}
	if c.OldQuizId != "" {
		m.Del(`<_uid_:` + c.OldQuizId + `> <quiz.candidate> <_uid_:` + c.Uid + `> .`)
		m.Del(`<_uid_:` + c.Uid + `> <candidate.quiz> <_uid_:` + c.OldQuizId + `> .`)
	}

	return m.String()
}

// TODO - Changing the quiz for a candidate doesn't work right now. Fix it.
func Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["id"]
	var c Candidate
	sr := server.Response{}
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var t time.Time
	if t, err = time.Parse(validityLayout, c.Validity); err != nil {
		sr.Message = "Couldn't parse the validity"
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	c.Uid = cid
	c.Validity = t.String()
	// TODO - Validate candidate fields shouldn't be empty.
	m := edit(c)
	res, err := dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if res.Code != "ErrorOk" {
		sr.Write(w, res.Message, "Mutation couldn't be applied by Dgraph.",
			http.StatusInternalServerError)
		return
	}
	go mail.Send(c.Name, c.Email, t.Format("Mon Jan 2 15:04:05 MST 2006"),
		c.Uid+c.Token)
	sr.Success = true
	sr.Message = "Candidate info updated successfully."
	w.Write(server.MarshalResponse(sr))
}

func get(candidateId string) string {
	return `
    {
	quiz.candidate(_uid_:` + candidateId + `) {
		name
		email
		token
		validity
		complete
		candidate.quiz {
			_uid_
			duration
		}
	  }
    }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["id"]
	q := get(cid)
	res, err := dgraph.Query(q)
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
