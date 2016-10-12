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
	Name     string
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
		  <_new_:qn> <name> "` + q.Name + `" .
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
			m += `
			<_new_:qn> <question.tag> <_uid_:` + t.Uid + `> .
			<_uid_:` + t.Uid + `> <tag.question> <_new_:qn> . `
		} else {
			m += `
			<_new_:t` + idx + `> <name> "` + t.Name + `" .
			<_new_:qn> <question.tag> <_new_:t` + idx + `> .
			<_new_:t` + idx + `> <tag.question> <_new_:qn> . `
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
	return m
}

func validateQuestion(q Question) error {
	if q.Name == "" || q.Text == "" {
		return fmt.Errorf("Question name/text can't be empty")
	}
	// TODO - Have validation on score.
	if q.Positive == 0 || q.Negative == 0 {
		return fmt.Errorf("Positive or negative score can't be zero.")
	}
	if len(q.Options) == 0 {
		return fmt.Errorf("Question should have atleast one option")
	}
	correct := 0
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correct++
		}
	}
	if correct == 0 {
		fmt.Errorf("Atleast one option should be correct")
	}
	return nil
}

// API for "Adding Question" to Database
func Add(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	sr := server.Response{}
	var q Question
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		sr.Error = "Couldn't decode JSON"
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	if err := validateQuestion(q); err != nil {
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	m := add(q)
	res := dgraph.SendMutation(m)
	if res.Code != "ErrorOk" {
		sr.Error = res.Message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.MarshalResponse(sr))
		return
	}

	sr.Success = true
	sr.Message = "Question Successfully Saved!"
	w.Write(server.MarshalResponse(sr))
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
		question_mutation = "{debug(_xid_: rootQuestion) { question (after: " + ques.Id + ", first: 10) { _uid_ name text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	} else {
		question_mutation = "{debug(_xid_: rootQuestion) { question (first:10) { _uid_ name text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
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
		  		name
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
	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + q.Uid + `> <name> "` + q.Name + `" .`)
	m.Set(`<_uid_:` + q.Uid + `> <text> "` + q.Text + `" .`)
	m.Set(`<_uid_:` + q.Uid + `> <positive> "` + strconv.FormatFloat(q.Positive, 'g', -1, 64) + `" .`)
	m.Set(`<_uid_:` + q.Uid + `> <negative> "` + strconv.FormatFloat(q.Negative, 'g', -1, 64) + `" .`)

	correct := 0
	for _, opt := range q.Options {
		m.Set(`<_uid_:` + opt.Uid + `> <name> "` + opt.Text + `" .`)
		m.Set(`<_uid_:` + q.Uid + `> <question.option> <_uid_:` + opt.Uid + `> . `)
		if opt.IsCorrect {
			correct++
			m.Set(`<_uid_:` + q.Uid + `> <question.correct> <_uid_:` + opt.Uid + `> .`)
		} else {
			m.Del(`<_uid_:` + q.Uid + `> <question.correct> <_uid_:` + opt.Uid + `> .`)
		}
	}

	// Create and associate Tags
	for i, t := range q.Tags {
		if t.Uid != "" && t.Is_delete {
			m.Del(`<_uid_:` + q.Uid + `> <question.tag> <_uid_:` + t.Uid + `> .`)
			m.Del(`<_uid_:` + t.Uid + `> <tag.question> <_uid_:` + q.Uid + `> . `)

		} else if t.Uid != "" {
			m.Set(`<_uid_:` + q.Uid + `> <question.tag> <_uid_:` + t.Uid + `> .`)
			m.Set(`<_uid_:` + t.Uid + `> <tag.question> <_uid_:` + q.Uid + `> . `)

		} else if t.Uid == "" {
			idx := strconv.Itoa(i)
			m.Set(`<_new_:tag` + idx + `> <name> "` + t.Name + `" .`)
			m.Set(`<_uid_:` + q.Uid + `> <question.tag> <_new_:tag` + idx + ` .`)
			m.Set(`<_new_:tag` + idx + `> <tag.question> <_uid_:` + q.Uid + `> . `)
		}
	}
	if correct > 1 {
		m.Set(`<_uid_:` + q.Uid + `> <multiple> "true" . `)
	} else {
		m.Set(`<_uid_:` + q.Uid + `> <multiple> "false" . `)
	}
	return m.String()
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
