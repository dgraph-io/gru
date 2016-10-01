package candidate

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/gorilla/mux"
)

type Candidate struct {
	Uid       string
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     string
	Validity  string `json:"validity"`
	QuizId    string `json:"quiz_id"`
	OldQuizId string `json:"old_quiz_id"`
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
	fmt.Println(q)
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
	fmt.Println(c)
	m := add(c)
	fmt.Println(m)
	res := dgraph.SendMutation(m)
	// TODO - Send a mail to the candidate with the link.
	if res.Success {
		res.Message = "Candidate added successfully."
	}
	server.WriteBody(w, res)
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
	fmt.Println(c)
	m := edit(c)
	fmt.Println(m)
	res := dgraph.SendMutation(m)
	if res.Success {
		res.Message = "Candidate info updated successfully."
	}
	server.WriteBody(w, res)
}

func get(candidateId string) string {
	return `
    {
        quiz.candidate(_uid_:` + candidateId + `) {
          name
          email
          validity
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
	fmt.Println(q)
	res := dgraph.Query(q)
	w.Write(res)
}
