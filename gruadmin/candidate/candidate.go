package candidate

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/mail"
	"github.com/dgraph-io/gru/gruadmin/server"
	quizp "github.com/dgraph-io/gru/gruserver/quiz"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type Candidate struct {
	Uid       string
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	Validity  string `json:"validity"`
	Complete  bool   `json:"complete,string"`
	QuizId    string `json:"quiz_id"`
	OldQuizId string `json:"old_quiz_id"`
	Quiz      []quiz `json:"candidate.quiz"`
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
        }
      }
    }
`
}

func Index(w http.ResponseWriter, r *http.Request) {
	quizId := r.URL.Query().Get("quiz_id")
	if quizId == "" {
		// TODO - Return error.
	}
	q := index(quizId)
	res := dgraph.Query(q)
	w.Write(res)
}

func add(c Candidate) string {
	// TODO - Add helper functions for sending mutations.
	return `
    mutation {
      set {
          <_uid_:` + c.QuizId + `> <quiz.candidate> <_new_:c> .
          <_new_:c> <candidate.quiz> <_uid_:` + c.QuizId + `> .
          <_new_:c> <email> "` + c.Email + `" .
          <_new_:c> <name> "` + c.Name + `" .
          <_new_:c> <token> "` + c.Token + `" .
          <_new_:c> <validity> "` + c.Validity + `" .
          <_new_:c> <complete> "false" .
	  }
    }`
}

func Add(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	var c Candidate
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		sr.Error = "Couldn't decode JSON"
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
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
	mr := dgraph.SendMutation(m)
	if mr.Code != "ErrorOk" {
		sr.Message = "Mutation couldn't be applied by Dgraph."
		sr.Error = mr.Message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.MarshalResponse(sr))
		return
	}

	// mutation applied successfully, lets send a mail to the candidate.
	uid, ok := mr.Uids["c"]
	if !ok {
		sr.Error = "Uid not returned for newly created candidate by Dgraph."
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.MarshalResponse(sr))
		return
	}

	// Token sent in mail is uid + the random string.
	go mail.Send(c.Name, c.Email, uid+c.Token)
	sr.Message = "Candidate added successfully."
	sr.Success = true
	w.Write(server.MarshalResponse(sr))
}

func edit(c Candidate) string {
	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + c.Uid + `> <email> "` + c.Email + `" . `)
	m.Set(`<_uid_:` + c.Uid + `> <name> "` + c.Name + `" . `)
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
	server.ReadBody(r, &c)

	sr := server.Response{}
	var t time.Time
	var err error
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
	res := dgraph.SendMutation(m)
	go mail.Send(c.Name, c.Email, c.Uid+c.Token)
	if res.Message == "ErrorOk" {
		sr.Success = true
		sr.Message = "Candidate info updated successfully."
	}
	server.WriteBody(w, sr)
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
	// TODO - Return error.
	if cid == "" {
	}
	q := get(cid)
	res := dgraph.Query(q)
	w.Write(res)
}

type quiz struct {
	Id        string           `json:"_uid_"`
	Duration  string           `json:"duration"`
	Questions []quizp.Question `json:"quiz.question"`
}

type qnIdsResp struct {
	Quizzes []quiz `json:"quiz"`
}

func quizQns(quizId string) []quizp.Question {
	q := `{
			quiz(_uid_: ` + quizId + `) {
				quiz.question {
				_uid_
				text
				positive
				negative
				question.option {
					_uid_
					name
				}
				multiple
			}
		}
	}`
	res := dgraph.Query(q)
	var resp qnIdsResp
	json.Unmarshal(res, &resp)
	if len(resp.Quizzes) != 1 {
		log.Fatal("Length of quizzes should just be 1")
	}
	return resp.Quizzes[0].Questions
}

type resp struct {
	Cand []Candidate `json:"quiz.candidate"`
}

type Res struct {
	Token    string `json:"token"`
	Duration string `json:"duration"`
}

func Validate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	sr := server.Response{}
	// This is the length of the random string. The id is uid + random string.
	if len(id) < 33 {
		w.WriteHeader(http.StatusUnauthorized)
		sr.Message = "Invalid token."
		w.Write(server.MarshalResponse(sr))
		return
	}

	// TODO - Check if the validity or the duration already elapsed.
	uid, token := id[:len(id)-33], id[len(id)-33:]

	c, err := quizp.ReadMap(uid)
	// Check for duplicate session.
	if err == nil && !c.LastExchange().IsZero() {
		timeSinceLastExchange := time.Now().Sub(c.LastExchange())
		// To avoid duplicate sessions.
		if timeSinceLastExchange < 10*time.Second {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Candidate doesn't exist in the map. So we get candidate info from uid and
	// insert it into map.
	q := get(uid)
	res := dgraph.Query(q)
	var resp resp
	json.Unmarshal(res, &resp)
	if len(resp.Cand) != 1 || len(resp.Cand[0].Quiz) != 1 {
		// No candidiate found with given uid
		w.WriteHeader(http.StatusUnauthorized)
		sr.Message = "Invalid token."
		w.Write(server.MarshalResponse(sr))
		return
	}

	if resp.Cand[0].Token != token || resp.Cand[0].Quiz[0].Id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		sr.Message = "Invalid token."
		w.Write(server.MarshalResponse(sr))
		return
	}

	var v time.Time
	if v, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", resp.Cand[0].Validity); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sr.Error = err.Error()
		w.Write(server.MarshalResponse(sr))
		return
	}

	if v.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		sr.Message = "Your token has already expired. Please contact contact@dgraph.io."
		w.Write(server.MarshalResponse(sr))
		return
	}

	if resp.Cand[0].Complete {
		w.WriteHeader(http.StatusUnauthorized)
		sr.Message = "You have already completed the quiz."
		w.Write(server.MarshalResponse(sr))
	}

	quiz := resp.Cand[0].Quiz[0]
	// Get quiz questions for the quiz id.
	qns := quizQns(quiz.Id)
	// TODO - Shuffle the order of questions.
	// x.Shuffle(ids)
	dur, err := time.ParseDuration(quiz.Duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sr.Error = err.Error()
		w.Write(server.MarshalResponse(sr))
	}

	quizp.New(uid, qns, dur)

	claims := x.Claims{
		UserId: uid,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := jwtToken.SignedString([]byte(*auth.Secret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sr.Error = err.Error()
		w.Write(server.MarshalResponse(sr))
		return
	}

	// TODO - Incase candidate already has a active session return error after
	// implementing Ping.
	// TODO - Also send quiz duration and time left incase candidate restarts.
	json.NewEncoder(w).Encode(Res{
		Token:    tokenString,
		Duration: dur.String(),
	})
}
