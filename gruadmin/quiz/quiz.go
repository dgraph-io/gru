package quiz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/x"
	"github.com/gorilla/mux"
)

type Quiz struct {
	Uid        string
	Name       string
	Duration   string
	Start_Date string
	End_Date   string
	Questions  []Question `json:"questions`
}

type Question struct {
	Uid       string `json:"_uid_"`
	Text      string
	Is_delete bool
}

func Add(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var quiz Quiz

	err := json.NewDecoder(r.Body).Decode(&quiz)
	if err != nil {
		panic(err)
	}
	//question_ids = quiz.Questions
	quiz_mutation := "mutation { set { <rootQuiz> <quiz> <_new_:quiz> . \n	<_new_:quiz> <name> \"" + quiz.Name +
		"\" . \n <_new_:quiz> <duration> \"" + quiz.Duration + "\" . \n"
	// "\" . \n <_new_:quiz> <start_date> \"" + quiz.Start_Date +
	// "\" . \n <_new_:quiz> <end_date> \"" + quiz.End_Date
	fmt.Println(quiz.Questions)
	for i := 0; i < len(quiz.Questions); i++ {
		quiz_mutation += "<_new_:quiz> <quiz.question> <_uid_:" + quiz.Questions[i].Uid + "> .\n"
	}
	quiz_mutation += " }}"
	x.Debug(quiz_mutation)
	_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
	if err != nil {
		panic(err)
	}
	stats := &server.Response{true, "Quiz Successfully Saved!"}
	quiz_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(quiz_json_response)
}

func Index(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	quiz_mutation := "{debug(_xid_: rootQuiz) { quiz { _uid_ name duration start_date end_date quiz.question { text }  }  }}"
	quiz_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
	if err != nil {
		panic(err)
	}
	defer quiz_response.Body.Close()
	quiz_body, err := ioutil.ReadAll(quiz_response.Body)
	if err != nil {
		panic(err)
	}
	x.Debug(string(quiz_body))

	jsonResp, err := json.Marshal(string(quiz_body))
	if err != nil {
		panic(err)
	}

	w.Write(jsonResp)
	w.Header().Set("Content-Type", "application/json")
}

// get quiz information

func get(quizId string) string {
	return `
    {
        root(_uid_:` + quizId + `) {
        	_uid_
        	name
        	duration
        	start_date
        	end_date
          quiz.question { _uid_ text }
        }
    }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	qid := vars["id"]
	// TODO - Return error.
	if qid == "" {
	}
	q := get(qid)
	x.Debug(q)
	res := dgraph.Query(q)
	w.Write(res)
}

// update quiz information

func edit(q Quiz) string {
	m := `
    mutation {
      set {
          <_uid_:` + q.Uid + `> <name> "` + q.Name + `" .
          <_uid_:` + q.Uid + `> <duration> "` + q.Duration + `" .`

	// Create and associate Tags
	for i := range q.Questions {
		if q.Questions[i].Is_delete == true {
			mutation := "mutation { delete { <_uid_:" + q.Uid + "> <quiz.question> <_uid_:" + q.Questions[i].Uid + "> .}}"
			x.Debug(mutation)
			_, err := http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(mutation))
			if err != nil {
				panic(err)
			}
		} else if q.Questions[i].Uid != "" {
			m += "\n<_uid_:" + q.Uid + "> <quiz.question> <_uid_:" + q.Questions[i].Uid + "> .\n"
		}
	}
	m += "}\n}"
	x.Debug(m)
	return m
}

func Edit(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	qid := vars["id"]
	// TODO - Return error.
	if qid == "" {
	}
	var q Quiz
	server.ReadBody(r, &q)
	q.Uid = qid
	// TODO - Validate candidate fields shouldn't be empty.
	x.Debug(q)
	m := edit(q)
	x.Debug(m)
	mr := dgraph.SendMutation(m)
	res := server.Response{}
	if mr.Code == "ErrorOk" {
		res.Success = true
		res.Message = "Quiz info updated successfully."
	}
	server.WriteBody(w, res)
}
