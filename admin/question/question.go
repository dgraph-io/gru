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
	Uid      string `json:"uid"`
	Name     string
	Text     string
	Positive float64
	Negative float64
	Notes    string
	Tags     []tag.Tag
	Options  []Option
}

type Option struct {
	Uid string `json:"uid"`
	// TODO - Change this to text later.
	Text      string `json:"name"`
	IsCorrect bool   `json:"is_correct"`
}

func add(q Question) *dgraph.Mutation {
	m := new(dgraph.Mutation)
	m.SetString("_:qn", "is_question", "")
	m.SetString("_:qn", "name", q.Name)
	m.SetString("_:qn", "text", q.Text)
	if q.Notes != "" {
		m.SetString("_:qn", "notes", q.Notes)
	}
	m.SetString("_:qn", "positive", strconv.FormatFloat(q.Positive, 'g', -1, 64))
	m.SetString("_:qn", "negative", strconv.FormatFloat(q.Negative, 'g', -1, 64))

	correct := 0
	for i, opt := range q.Options {
		idx := strconv.Itoa(i)
		optKey := "_:o" + idx
		m.SetLink("_:qn", "question.option", optKey)
		m.SetString(optKey, "name", opt.Text)
		if opt.IsCorrect {
			m.SetLink("_:qn", "question.correct", optKey)
			correct++
		}
	}

	for i, t := range q.Tags {
		idx := strconv.Itoa(i)
		if t.Uid != "" {
			m.SetLink("_:qn", "question.tag", t.Uid)
		} else {
			tagKey := "_:t" + idx
			m.SetString(tagKey, "is_tag", "")
			m.SetString(tagKey, "name", t.Name)
			m.SetLink("_:qn", "question.tag", tagKey)
		}
	}

	if correct > 1 {
		m.SetString("_:qn", "multiple", "true")
	} else {
		m.SetString("_:qn", "multiple", "false")
	}
	return m
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
		return fmt.Errorf("Question should have at least one option")
	}
	correct := 0
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correct++
		}
		if opt.Text == "" {
			return fmt.Errorf("Option text can't be empty.")
		}
	}
	if correct == 0 {
		return fmt.Errorf("At least one option should be correct")
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
	query := `{
		questions(func: has(is_question)) {
			uid
			name
			text
			negative
			positive
			notes
			question.tag {
				uid
				name
			}
			question.option {
				uid
				name
			}
			question.correct {
				uid
				name
			}
		}
	}`

	b, err := dgraph.Query(query)
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func getQuestionQuery(questionId string) string {
	return `
	{
		question(func: uid(` + questionId + `)) {
			uid
			name
			text
			positive
			negative
			notes
			question.option	{
				uid
				name
			}
			question.correct {
				uid
				name
			}
			question.tag {
				uid
				name
			}
		}
	}`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qid := vars["id"]

	res, err := dgraph.Query(getQuestionQuery(qid))
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	} // TODO - Check if Dgraph returns record not found and wrap it with an error.
	w.Write(res)
}

// update question
func edit(q Question) (*dgraph.Mutation, error) {
	if err := validateQuestion(q); err != nil {
		return nil, err
	}

	m := new(dgraph.Mutation)

	m.SetString(q.Uid, "name", q.Name)
	m.SetString(q.Uid, "text", q.Text)
	if q.Notes != "" {
		m.SetString(q.Uid, "notes", q.Notes)
	}

	m.SetString(q.Uid, "positive", strconv.FormatFloat(q.Positive, 'g', -1, 64))
	m.SetString(q.Uid, "negative", strconv.FormatFloat(q.Negative, 'g', -1, 64))

	correct := 0
	for _, opt := range q.Options {
		if opt.Text == "" {
			return nil, fmt.Errorf("Option text can't be empty.")
		}
		m.SetString(opt.Uid, "name", opt.Text)
		m.SetLink(q.Uid, "question.option", opt.Uid)
		if opt.IsCorrect {
			correct++
			m.SetLink(q.Uid, "question.correct", opt.Uid)
		} else {
			m.DelLink(q.Uid, "question.correct", opt.Uid)
		}
	}

	tagCount := 0
	// Create and associate Tags
	for i, t := range q.Tags {
		if t.Uid != "" && t.Is_delete {
			m.DelLink(q.Uid, "question.tag", t.Uid)
		} else if t.Uid != "" {
			tagCount++
			m.SetLink(q.Uid, "question.tag", t.Uid)
		} else if t.Uid == "" {
			if t.Name == "" {
				return nil, fmt.Errorf("Tag name can't be empty")
			}
			tagCount++
			tagKey := "_:tag" + strconv.Itoa(i)
			m.SetString(tagKey, "name", t.Name)
			m.SetString(tagKey, "is_tag", "")
			m.SetLink(q.Uid, "question.tag", tagKey)
		}
	}
	if tagCount == 0 {
		return nil, fmt.Errorf("A question should have at least one tag")
	}

	if correct == 0 {
		return nil, fmt.Errorf("At least one option should be correct")
	} else if correct > 1 {
		m.SetString(q.Uid, "multiple", "true")
	} else {
		m.SetString(q.Uid, "multiple", "false")
	}
	return m, nil
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

	mutation, err := edit(q);
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusBadRequest)
		return
	}

	mr, err := dgraph.SendMutation(mutation)
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
