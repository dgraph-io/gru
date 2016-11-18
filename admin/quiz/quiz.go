package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/gorilla/mux"
)

type Quiz struct {
	Uid       string
	Name      string
	Duration  int
	Cutoff    float64    `json:"cut_off"`
	Questions []Question `json:"questions`
}

type Question struct {
	Uid       string `json:"_uid_"`
	Text      string
	Is_delete bool
}

func Add(w http.ResponseWriter, r *http.Request) {
	var q Quiz
	sr := server.Response{}
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		sr.Write(w, "Couldn't decode JSON", "", http.StatusBadRequest)
		return
	}

	m := new(dgraph.Mutation)
	m.Set(`<root> <quiz> <_new_:quiz> .`)
	// TODO - Error if Name is empty.
	m.Set(`<_new_:quiz> <name> "` + q.Name + `" .`)
	m.Set(`<_new_:quiz> <cut_off> "` + strconv.FormatFloat(q.Cutoff, 'g', -1, 64) + `" .`)
	m.Set(`<_new_:quiz> <duration> "` + strconv.Itoa(q.Duration) + `" . `)
	for _, q := range q.Questions {
		m.Set(`<_new_:quiz> <quiz.question> <_uid_:` + q.Uid + `> .`)
	}

	mr, err := dgraph.SendMutation(m.String())
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if mr.Code != "ErrorOk" {
		sr.Write(w, mr.Message, "", http.StatusInternalServerError)
		return
	}

	sr.Success = true
	sr.Message = "Quiz Successfully Saved!"
	w.Write(server.MarshalResponse(sr))
}

func Index(w http.ResponseWriter, r *http.Request) {
	q := "{debug(_xid_: root) { quiz { _uid_ name duration quiz.question { text }  }  }}"
	res, err := dgraph.Query(q)
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

// get quiz information

func get(quizId string) string {
	return `
    {
	root(_uid_:` + quizId + `) {
		_uid_
		name
		duration
		cut_off
		quiz.question { _uid_ name text positive negative question.tag { _uid_ name } question.correct { _uid_ name}}
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
	}
	w.Write(res)
}

func edit(q Quiz) string {
	m := new(dgraph.Mutation)
	// TODO - Validate these fields.
	m.Set(`<_uid_:` + q.Uid + `> <name> "` + q.Name + `" .`)
	m.Set(`<_uid_:` + q.Uid + `> <duration> "` + strconv.Itoa(q.Duration) + `" .`)
	m.Set(`<_uid_:` + q.Uid + `> <cut_off> "` + strconv.FormatFloat(q.Cutoff, 'g', -1, 64) + `" .`)

	// Create and associate Tags
	for _, que := range q.Questions {
		if que.Is_delete {
			m.Del(`<_uid_:` + q.Uid + `> <quiz.question> <_uid_:` + que.Uid + `> .`)
		} else if que.Uid != "" {
			m.Set(`<_uid_:` + q.Uid + `> <quiz.question> <_uid_:` + que.Uid + `> . `)
		}
	}
	return m.String()
}

func Edit(w http.ResponseWriter, r *http.Request) {
	var q Quiz
	vars := mux.Vars(r)
	qid := vars["id"]
	sr := server.Response{}
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		sr.Write(w, "Couldn't decode JSON", "", http.StatusBadRequest)
		return
	}
	q.Uid = qid
	// TODO - Validate candidate fields shouldn't be empty.
	m := edit(q)
	_, err = dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}

	sr.Success = true
	sr.Message = "Quiz info updated successfully."
	w.Write(server.MarshalResponse(sr))
}
