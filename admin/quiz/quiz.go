package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/gorilla/mux"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

type Quiz struct {
	Uid       string
	Name      string
	Duration  int
	Cutoff    float64    `json:"cut_off"`
	Threshold float64    `json:"threshold"`
	Questions []Question `json:"questions"`
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
	m.Set(`<root> <quiz> <_:quiz> .`)
	// TODO - Error if Name is empty.
	m.Set(`<_:quiz> <name> "` + q.Name + `" .`)
	m.Set(`<_:quiz> <threshold> "` + strconv.FormatFloat(q.Threshold, 'g', -1, 64) + `" .`)
	m.Set(`<_:quiz> <cut_off> "` + strconv.FormatFloat(q.Cutoff, 'g', -1, 64) + `" .`)
	m.Set(`<_:quiz> <duration> "` + strconv.Itoa(q.Duration) + `" . `)
	for _, q := range q.Questions {
		m.Set(`<_:quiz> <quiz.question> <` + q.Uid + `> .`)
	}

	mr, err := dgraph.SendMutation(m.String())
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if mr.Code != dgraph.Success {
		sr.Write(w, mr.Message, "", http.StatusInternalServerError)
		return
	}

	sr.Success = true
	sr.Message = "Quiz Successfully Saved!"
	w.Write(server.MarshalResponse(sr))
}

func Index(w http.ResponseWriter, r *http.Request) {
	q := `{
		debug(id: root) {
			quiz {
				_uid_
				name
				duration
				quiz.question {
					_uid_
					text
				}
			}
		}
	}`
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
	root(id:` + quizId + `) {
		_uid_
		name
		duration
		cut_off
		threshold
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
	m.Set(`<` + q.Uid + `> <name> "` + q.Name + `" .`)
	m.Set(`<` + q.Uid + `> <duration> "` + strconv.Itoa(q.Duration) + `" .`)
	m.Set(`<` + q.Uid + `> <threshold> "` + strconv.FormatFloat(q.Threshold, 'g', -1, 64) + `" .`)
	m.Set(`<` + q.Uid + `> <cut_off> "` + strconv.FormatFloat(q.Cutoff, 'g', -1, 64) + `" .`)

	// Create and associate Tags
	for _, que := range q.Questions {
		if que.Is_delete {
			m.Del(`<` + q.Uid + `> <quiz.question> <` + que.Uid + `> .`)
		} else if que.Uid != "" {
			m.Set(`<` + q.Uid + `> <quiz.question> <` + que.Uid + `> . `)
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
