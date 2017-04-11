package question

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/admin/tag"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/gorilla/mux"
)

type Question struct {
	Uid      string `json:"_uid_"`
	Name     string
	Text     string
	Positive float64
	Negative float64
	Notes    string
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
	m := new(dgraph.Mutation)
	m.Set(`<root> <question> <_:qn> .`)
	m.Set(`<_:qn> <name> "` + q.Name + `" .`)
	m.Set(`<_:qn> <text> "` + q.Text + `" .`)
	if q.Notes != "" {
		m.Set(`<_:qn> <notes> "` + q.Notes + `" .`)
	}
	m.Set(`<_:qn> <positive> "` + strconv.FormatFloat(q.Positive, 'g', -1, 64) + `" .`)
	m.Set(`<_:qn> <negative> "` + strconv.FormatFloat(q.Negative, 'g', -1, 64) + `" .`)

	correct := 0
	for i, opt := range q.Options {
		idx := strconv.Itoa(i)
		m.Set(`<_:qn> <question.option> <_:o` + idx + `> .`)
		m.Set(`<_:o` + idx + `> <name> "` + opt.Text + `" .`)
		if opt.IsCorrect {
			m.Set(`<_:qn> <question.correct> <_:o` + idx + `> .`)
			correct++
		}
	}

	for i, t := range q.Tags {
		idx := strconv.Itoa(i)
		if t.Uid != "" {
			m.Set(`<_:qn> <question.tag> <` + t.Uid + `> .`)
			m.Set(`<` + t.Uid + `> <tag.question> <_:qn> . `)
		} else {
			m.Set(`<_:t` + idx + `> <name> "` + t.Name + `" .`)
			m.Set(`<_:qn> <question.tag> <_:t` + idx + `> .`)
			m.Set(`<_:t` + idx + `> <tag.question> <_:qn> .`)
		}
	}

	if correct > 1 {
		m.Set(`<_:qn> <multiple> "true" . `)
	} else {
		m.Set(`<_:qn> <multiple> "false" . `)
	}
	return m.String()
}

// TODO - Move this inline with add, like we have for edit.
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
		return fmt.Errorf("Atleast one option should be correct")
	}
	return nil
}

// API for "Adding Question" to Database
func Add(w http.ResponseWriter, r *http.Request) {
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
	res, err := dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if res.Code != dgraph.Success {
		sr.Write(w, res.Message, "", http.StatusInternalServerError)
		return
	}

	sr.Success = true
	sr.Message = "Question Successfully Saved!"
	w.Write(server.MarshalResponse(sr))
}

type qid struct {
	Id string
}

func Index(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	var q qid
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		sr.Write(w, "", "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var query string
	var after string
	if q.Id == "" {
		after = "0"
	} else {
		after = q.Id
	}

	query = `{
		debug(id: root) {
			question (after:` + after + `, first: 20) {
				_uid_
				name
				text
				negative
				positive
				notes
				question.tag {
					_uid_
					name
				}
				question.option {
					name
				}
				question.correct {
					name
				}
			}
		}
	}`

	b, err := dgraph.Query(query)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

// get question information
func get(questionId string) string {
	return `
	{
		root(id:` + questionId + `) {
			_uid_
			name
			text
			positive
			negative
			notes
			question.option	{
				_uid_
				name
			}
			question.correct {
				_uid_
				name
			}
			question.tag {
				_uid_
				name
			}
		}
	}`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qid := vars["id"]

	q := get(qid)
	res, err := dgraph.Query(q)
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	} // TODO - Check if Dgraph returns record not found and wrap it with an error.
	w.Write(res)
}

// update question
func edit(q Question) (string, error) {
	m := new(dgraph.Mutation)
	if q.Name == "" || q.Text == "" {
		return "", fmt.Errorf("Question name/text can't be empty.")
	}
	m.Set(`<` + q.Uid + `> <name> "` + q.Name + `" .`)
	m.Set(`<` + q.Uid + `> <text> "` + q.Text + `" .`)
	if q.Notes != "" {
		m.Set(`<` + q.Uid + `> <notes> "` + q.Notes + `" .`)
	}

	if q.Positive == 0 || q.Negative == 0 {
		return "", fmt.Errorf("Positive/Negative score can't be zero.")
	}
	m.Set(`<` + q.Uid + `> <positive> "` + strconv.FormatFloat(q.Positive, 'g', -1, 64) + `" .`)
	m.Set(`<` + q.Uid + `> <negative> "` + strconv.FormatFloat(q.Negative, 'g', -1, 64) + `" .`)

	correct := 0
	if len(q.Options) == 0 {
		return "", fmt.Errorf("Question should have atleast one option")
	}
	for _, opt := range q.Options {
		if opt.Text == "" {
			return "", fmt.Errorf("Option text can't be empty.")
		}
		m.Set(`<` + opt.Uid + `> <name> "` + opt.Text + `" .`)
		m.Set(`<` + q.Uid + `> <question.option> <` + opt.Uid + `> . `)
		if opt.IsCorrect {
			correct++
			m.Set(`<` + q.Uid + `> <question.correct> <` + opt.Uid + `> .`)
		} else {
			m.Del(`<` + q.Uid + `> <question.correct> <` + opt.Uid + `> .`)
		}
	}

	// Create and associate Tags
	for i, t := range q.Tags {
		if t.Uid != "" && t.Is_delete {
			m.Del(`<` + q.Uid + `> <question.tag> <` + t.Uid + `> .`)
			m.Del(`<` + t.Uid + `> <tag.question> <` + q.Uid + `> . `)

		} else if t.Uid != "" {
			m.Set(`<` + q.Uid + `> <question.tag> <` + t.Uid + `> .`)
			m.Set(`<` + t.Uid + `> <tag.question> <` + q.Uid + `> . `)

		} else if t.Uid == "" {
			if t.Name == "" {
				return "", fmt.Errorf("Tag name can't be empty.")
			}
			idx := strconv.Itoa(i)
			m.Set(`<_:tag` + idx + `> <name> "` + t.Name + `" .`)
			m.Set(`<` + q.Uid + `> <question.tag> <_:tag` + idx + `> .`)
			m.Set(`<_:tag` + idx + `> <tag.question> <` + q.Uid + `> . `)
		}
	}
	// TODO - There should be atleast one tag associated with a question.

	if correct == 0 {
		return "", fmt.Errorf("Atleast one option should be correct.")
	} else if correct > 1 {
		m.Set(`<` + q.Uid + `> <multiple> "true" . `)
	} else {
		m.Set(`<` + q.Uid + `> <multiple> "false" . `)
	}
	return m.String(), nil
}

func Edit(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	// vars := mux.Vars(r)
	// qid := vars["id"]
	// TODO - Id should be obtained from url not the body.
	var q Question
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		sr.Write(w, "", "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var m string
	if m, err = edit(q); err != nil {
		sr.Write(w, "", err.Error(), http.StatusBadRequest)
		return
	}

	mr, err := dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if mr.Code != dgraph.Success {
		sr.Write(w, mr.Message, "", http.StatusInternalServerError)
		return
	}

	sr.Success = true
	sr.Message = "Question updated successfully."
	w.Write(server.MarshalResponse(sr))
}
