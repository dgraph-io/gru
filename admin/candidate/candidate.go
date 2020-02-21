package candidate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/gorilla/mux"
)

type Candidate struct {
	Uid       string `json:"uid"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	Validity  string `json:"validity"`
	QuizId    string `json:"quiz_id"`
	OldQuizId string `json:"old_quiz_id"`
}

const (
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
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
	quiz(func: uid(` + quizId + `)) {
		quiz.candidate {
			uid
			name
			email
			score
			token
			validity
			complete
			deleted
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

func AddCand(quizId, name, email string, validity time.Time) (string, error) {
	m := new(dgraph.Mutation)
	token := randStringBytes(33)
	// TODO: Use reverse predicate for quiz.candidate & candidate.quiz
	m.SetLink(quizId, "quiz.candidate", "_:c")
	m.SetLink("_:c", "candidate.quiz", quizId)
	m.SetString("_:c", "email", email)
	m.SetString("_:c", "name", name)
	m.SetString("_:c", "token", token)
	m.SetString("_:c", "validity", validity.Format(time.RFC3339Nano))
	m.SetString("_:c", "invite_sent", time.Now().UTC().String())
	m.SetString("_:c", "complete", "false")

	mr, err := dgraph.SendMutation(m)
	if err != nil {
		return "", err
	}

	// mutation applied successfully, lets send a mail to the candidate.
	uid, ok := mr.Uids["c"]
	if !ok {
		return "", fmt.Errorf("Uid not returned for newly created candidate.")

	}

	// Token sent in mail is uid + the random string.
	go mail.Send(email, validity.Format("dd MMMM yyyy"), uid+token)
	return uid, nil
}

type addCand struct {
	Emails   []string
	Validity string
	QuizId   string `json:"quiz_id"`
}

func Add(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	var c addCand
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var t time.Time
	if t, err = time.Parse(time.RFC3339Nano, c.Validity); err != nil {
		sr.Write(w, err.Error(), "Couldn't parse the validity", http.StatusBadRequest)
		return
	}

	for _, email := range c.Emails {
		if _, err := AddCand(c.QuizId, "", email, t); err != nil {
			sr.Write(w, err.Error(), "", http.StatusInternalServerError)
			return
		}
	}
	sr.Success = true
	sr.Message = "Candidates invited successfully."
	w.Write(server.MarshalResponse(sr))
}

func edit(c Candidate) *dgraph.Mutation {
	m := new(dgraph.Mutation)
	m.SetString(c.Uid, "email", c.Email)
	m.SetString(c.Uid, "validity", c.Validity)

	// When the quiz for which candidate is invited is changed, we get both OldQuizId
	// and new QuizId.
	if c.QuizId != "" {
		m.SetLink(c.QuizId, "quiz.candidate", c.Uid)
		m.SetLink(c.Uid, "candidate.quiz", c.QuizId)
	}
	if c.OldQuizId != "" {
		m.DelLink(c.OldQuizId, "quiz.candidate", c.Uid)
		m.DelLink(c.Uid, "candidate.quiz", c.OldQuizId)
	}

	return m
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

	t, err := time.Parse(time.RFC3339Nano, c.Validity);
	if err != nil {
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
	if res.Code != dgraph.Success {
		sr.Write(w, res.Message, "Mutation couldn't be applied.",
			http.StatusInternalServerError)
		return
	}
	sr.Success = true
	sr.Message = "Candidate info updated successfully."
	w.Write(server.MarshalResponse(sr))
}

func get(candidateId string) string {
	return `{
		quiz.candidate(func: uid(` + candidateId + `)) {
			name
			email
			token
			validity
			complete
			candidate.quiz {
				uid
				duration
			}
	  }
  }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	res, err := dgraph.Query(get(vars["id"]))
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

type resendReq struct {
	Email    string
	Token    string
	Validity string
}

func ResendInvite(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	cid := mux.Vars(r)["id"]

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusBadRequest)
		return
	}
	var rr resendReq
	if err := json.Unmarshal(b, &rr); err != nil {
		sr.Write(w, "", err.Error(), http.StatusBadRequest)
		return
	}

	if rr.Email == "" || rr.Token == "" || rr.Validity == "" {
		sr.Write(w, "", "Email/token/validity can't be empty.", http.StatusBadRequest)
		return
	}

	t, err := time.Parse(time.RFC3339Nano, rr.Validity);
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't parse the validity", http.StatusBadRequest)
		return
	}

	go mail.Send(rr.Email, t.Format("dd MMMM yyyy"), cid+rr.Token)

	sr.Success = true
	sr.Write(w, "", "Invite has been resent.", http.StatusOK)
}

type candInfo struct {
	Data struct {
		Candidates []Candidate
	}
}

func candName(id string) string {
	q := `query {
    candidate(func: uid(` + id + `)) {
      name
    }
  }`
	var ci candInfo
	if err := dgraph.QueryAndUnmarshal(q, &ci); err != nil {
		return ""
	}
	if len(ci.Data.Candidates) != 1 {
		return ""
	}
	return ci.Data.Candidates[0].Name
}
