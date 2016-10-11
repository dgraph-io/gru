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
	QuizId    string `json:"quiz_id"`
	OldQuizId string `json:"old_quiz_id"`
	Quiz      []quiz `json:"candidate.quiz"`
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

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
        }
      }
    }
`
}

func Index(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	quizId := r.URL.Query().Get("quiz_id")
	if quizId == "" {
		// TODO - Return error.
	}
	q := index(quizId)
	x.Debug(q)
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
      }
    }`
}

func Add(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var c Candidate
	server.ReadBody(r, &c)
	// TODO - Validate candidate fields shouldn't be empty.
	c.Token = randStringBytes(33)
	x.Debug(c)
	m := add(c)
	x.Debug(m)
	res := dgraph.SendMutation(m)
	sr := server.Response{}
	if res.Code != "ErrorOk" {
		sr.Message = "Mutation couldn't be applied by Dgraph."
		server.WriteBody(w, sr)
		return
	}
	// mutation applied successfully, lets send a mail to the candidate.
	uid, ok := res.Uids["c"]
	if !ok {
		log.Fatal("Uid not returned for newly created candidate by Dgraph.")
	}
	x.Debug(uid)
	// Token sent in mail is uid + the random string.
	// TODO - Move this to a background goroutine.
	go mail.Send(c.Name, c.Email, uid+c.Token)
	sr.Message = "Candidate added successfully."
	sr.Success = true
	server.WriteBody(w, sr)
}

func edit(c Candidate) string {
	// TODO - Handler changing quiz_id
	// var del string
	// if c.OldQuizId != "" {
	// 	del = `
	//     delete {
	//         <_uid_:` + c.OldQuizId + `> <quiz.candidate> <_uid_:` + c.Id + `> .
	//         <_uid_:` + c.Id + `> <candidate.quiz> <_uid_:` + c.OldQuizId + `> .
	//     }`
	// }
	m := `
    mutation {
      set {
          <_uid_:` + c.Uid + `> <email> "` + c.Email + `" .
          <_uid_:` + c.Uid + `> <name> "` + c.Name + `" .
          <_uid_:` + c.Uid + `> <validity> "` + c.Validity + `" .
      }
    }`
	return m
}

func Edit(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	cid := vars["id"]
	// TODO - Return error.
	if cid == "" {
	}
	var c Candidate
	server.ReadBody(r, &c)
	c.Uid = cid
	// TODO - Validate candidate fields shouldn't be empty.
	x.Debug(c)
	m := edit(c)
	x.Debug(m)
	res := dgraph.SendMutation(m)
	sr := server.Response{}
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
          candidate.quiz {
		    _uid_
		    duration
		  }
	  }
    }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	cid := vars["id"]
	// TODO - Return error.
	if cid == "" {
	}
	q := get(cid)
	x.Debug(q)
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
	x.Debug(string(res))
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
	Token string `json:"token"`
}

func Validate(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	// This is the length of the random string. The id is uid + random string.
	if len(id) < 33 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if resp.Cand[0].Token != token || resp.Cand[0].Quiz[0].Id == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	quiz := resp.Cand[0].Quiz[0]
	// Get quiz questions for the quiz id.
	qns := quizQns(quiz.Id)
	// x.Shuffle(ids)
	// TODO - Although we verify the duration when the quiz is created, still handle
	// error here.
	dur, _ := time.ParseDuration(quiz.Duration)
	quizp.New(uid, qns, dur)

	// TODO - Check token validity.
	claims := x.Claims{
		UserId: uid,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := jwtToken.SignedString([]byte(*auth.Secret))
	if err != nil {
		log.Fatal(err)
	}
	x.Debug(tokenString)

	// TODO - Incase candidate already has a active session return error after
	// implementing Ping.
	// TODO - Also send quiz duration and time left incase candidate restarts.
	json.NewEncoder(w).Encode(Res{Token: tokenString})
}
