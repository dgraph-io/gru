package question

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/gruadmin/tag"
	"github.com/dgraph-io/gru/x"
	"github.com/gorilla/mux"
)

type Question struct {
	Uid      string `json:"_uid_"`
	Text     string
	Positive float64
	Negative float64
	Tags     []tag.Tag
	Options  []Option
}

type Option struct {
	Uid string `json:"_uid_"`
	// TODO - Change this to text later.
	Text      string `json:"name"`
	IsCorrect bool   `json:"is_correct"`
}

func add(q Question) string {
	m := `mutation {
		set {
		  <rootQuestion> <question> <_new_:qn> .
		  <_new_:qn> <text> "` + q.Text + `" .
		  <_new_:qn> <positive> "` + strconv.FormatFloat(q.Positive, 'g', -1, 64) + `" .
		  <_new_:qn> <negative> "` + strconv.FormatFloat(q.Negative, 'g', -1, 64) + `" .`

	correct := 0
	for i, opt := range q.Options {
		idx := strconv.Itoa(i)
		m += `
		<_new_:qn> <question.option> <_new_:o` + idx + `> .
		<_new_:o` + idx + `> <name> "` + opt.Text + `" .`
		if opt.IsCorrect {
			m += `
			<_new_:qn> <question.correct> <_new_:o` + idx + `> .`
			correct++
		}
	}

	for i, t := range q.Tags {
		idx := strconv.Itoa(i)
		if t.Uid != "" {
			x.Debug(t.Uid)
			m += `
			<_new_:qn> <question.tag> <_uid_:` + t.Uid + `> .
			<_uid_:` + t.Uid + `> <tag.question> <_new_:qn> . `
		} else {
			m += `
			<_new_:t` + idx + `> <name> "` + t.Name + `" .
			<_new_:qn> <question.tag> <_new_:tag` + idx + `> .
			<_new_:tag` + idx + `> <tag.question> <_new_:qn> . `
		}
	}

	if correct > 1 {
		m += `
		<_new_:qn> <multiple> "true" . `
	} else {
		m += `
		<_new_:qn> <multiple> "false" . `
	}
	m += `
	  }
  }	`
	fmt.Println(m)
	return m
}

// API for "Adding Question" to Database
func Add(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	var ques Question

	if r.Method == "OPTIONS" {
		return
	}
	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		log.Fatal(err)
	}
	x.Debug(ques)

	m := add(ques)
	res := dgraph.SendMutation(m)
	sr := server.Response{}
	if res.Code == "ErrorOk" {
		sr.Success = true
		sr.Message = "Question Successfully Saved!"
	}
	question_json_response, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.Write(question_json_response)
}

// FETCH All Questions HANDLER: Incomplete
func Index(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var ques GetQuestion
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	x.Debug(ques)
	var question_mutation string
	if ques.Id != "" {
		question_mutation = "{debug(_xid_: rootQuestion) { question (after: " + ques.Id + ", first: 10) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	} else {
		question_mutation = "{debug(_xid_: rootQuestion) { question (first:10) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	}
	x.Debug(question_mutation)
	w.Header().Set("Content-Type", "application/json")

	question_response, err := http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(question_mutation))
	if err != nil {
		panic(err)
	}
	defer question_response.Body.Close()
	question_body, err := ioutil.ReadAll(question_response.Body)
	if err != nil {
		panic(err)
	}
	x.Debug(string(question_body))

	jsonResp, err := json.Marshal(string(question_body))
	if err != nil {
		panic(err)
	}

	w.Write(jsonResp)
}

type QuestionAPIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	UIDS    struct {
		Question uint64 `json:question`
	}
}

// method to parse question response
func parseQuestionResponse(question_body []byte) (*QuestionAPIResponse, error) {
	var question_response = new(QuestionAPIResponse)
	err := json.Unmarshal(question_body, &question_response)
	if err != nil {
		log.Fatal(err)
	}
	return question_response, err
}

type TagFilter struct {
	UID string
}

type GetQuestion struct {
	Id string
}

// FILTER QUESTION HANDLER: Fileter By Tags
func Filter(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var tag TagFilter

	err := json.NewDecoder(r.Body).Decode(&tag)
	if err != nil {
		panic(err)
	}

	filter_query := "{root(_uid_: " + tag.UID + ") { tag.question { text }}"
	filter_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(filter_query))
	if err != nil {
		panic(err)
	}
	defer filter_response.Body.Close()
	filter_body, err := ioutil.ReadAll(filter_response.Body)
	if err != nil {
		panic(err)
	}
	x.Debug(string(filter_body))
	jsonResp, err := json.Marshal(string(filter_body))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

// get question information

func get(questionId string) string {
	return `
    {
        root(_uid_:` + questionId + `) {
		  _uid_
          text
          positive
          negative
          question.option { _uid_ name }
          question.correct { _uid_ name }
          question.tag { _uid_ name }
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

// update question

func edit(q Question) string {
	m := `
    mutation {
      set {
          <_uid_:` + q.Uid + `> <text> "` + q.Text + `" .
          <_uid_:` + q.Uid + `> <positive> "` + strconv.FormatFloat(q.Positive, 'g', -1, 64) + `" .
          <_uid_:` + q.Uid + `> <negative> "` + strconv.FormatFloat(q.Negative, 'g', -1, 64) + `" .`

	correct := 0
	for l := range q.Options {
		m += "<_uid_:" + q.Options[l].Uid + "> <name> \"" + q.Options[l].Text +
			"\" .\n <_uid_:" + q.Uid + "> <question.option> <_uid_:" + q.Options[l].Uid + "> . \n "

		if q.Options[l].IsCorrect == true {
			correct++
			m += "<_uid_:" + q.Uid + "> <question.correct> <_uid_:" + q.Options[l].Uid + "> . \n "
		}
		if q.Options[l].IsCorrect == false {
			delete_correct := "mutation { delete { <_uid_:" + q.Uid + "> <question.correct> <_uid_:" + q.Options[l].Uid + "> .}}"
			_, err := http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(delete_correct))
			if err != nil {
				panic(err)
			}
		}
	}

	// Create and associate Tags
	for i := range q.Tags {
		if q.Tags[i].Uid != "" && q.Tags[i].Is_delete == true {
			query_mutation := "mutation { delete { <_uid_:" + q.Uid + "> <question.tag> <_uid_:" + q.Tags[i].Uid +
				"> .\n <_uid_:" + q.Tags[i].Uid + "> <tag.question> <_uid_:" + q.Uid + "> . \n }}"
			x.Debug(query_mutation)
			dgraph.SendMutation(query_mutation)

		} else if q.Tags[i].Uid != "" {
			m += "<_uid_:" + q.Uid + "> <question.tag> <_uid_:" + q.Tags[i].Uid +
				"> . \n <_uid_:" + q.Tags[i].Uid + "> <tag.question> <_uid_:" + q.Uid + "> . \n "
		} else if q.Tags[i].Uid == "" {
			index := strconv.Itoa(i)
			m += "<_new_:tag" + index + "> <name> \"" + q.Tags[i].Name +
				"\" .\n <_uid_:" + q.Uid + "> <question.tag> <_new_:tag" + index +
				"> . \n <_new_:tag" + index + "> <tag.question> <_uid_:" + q.Uid + "> . \n "
		}
	}
	if correct > 1 {
		m += "<_uid_:" + q.Uid + "> <multiple> \"true\" . \n"
	} else {
		m += "<_uid_:" + q.Uid + "> <multiple> \"false\" . \n"
	}

	m += " }}"
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
	var q Question
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
		res.Message = "Question info updated successfully."
	}
	server.WriteBody(w, res)
}
