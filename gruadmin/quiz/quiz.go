package quiz

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
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
	var q Quiz
	sr := server.Response{}
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		sr.Error = "Couldn't decode JSON"
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	// We should be able to parse the duration.
	_, err = time.ParseDuration(q.Duration)
	if err != nil {
		sr.Message = "Couldn't parse duration."
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	m := new(dgraph.Mutation)
	m.Set(`<rootQuiz> <quiz> <_new_:quiz> .`)
	// TODO - Error if Name is empty.
	m.Set(`<_new_:quiz> <name> "` + q.Name + `" .`)
	m.Set(`<_new_:quiz> <duration> "` + q.Duration + `" . `)
	for _, q := range q.Questions {
		m.Set(`<_new_:quiz> <quiz.question> <_uid_:` + q.Uid + `> .`)
	}

	mr := dgraph.SendMutation(m.String())
	if mr.Code != "ErrorOk" {
		sr.Error = mr.Message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.MarshalResponse(sr))
		return
	}

	sr.Success = true
	sr.Message = "Quiz Successfully Saved!"
	w.Write(server.MarshalResponse(sr))
}

func Index(w http.ResponseWriter, r *http.Request) {
	q := "{debug(_xid_: rootQuiz) { quiz { _uid_ name duration start_date end_date quiz.question { text }  }  }}"
	res := dgraph.Query(q)
	// TODO - Remove this, sent byte slice directly.
	jsonResp, _ := json.Marshal(string(res))
	w.Write(jsonResp)
}

// get quiz information

func get(quizId string) string {
	return `
    {
        root(_uid_:` + quizId + `) {
        	_uid_
        	name
        	duration
          quiz.question { _uid_ name text positive negative question.tag { _uid_ name }}
        }
    }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qid := vars["id"]
	q := get(qid)
	res := dgraph.Query(q)
	w.Write(res)
}

func edit(q Quiz) string {
	m := new(dgraph.Mutation)
	// TODO - Validate these fields.
	m.Set(`<_uid_:` + q.Uid + `> <name> "` + q.Name + `" .`)
	m.Set(`<_uid_:` + q.Uid + `> <duration> "` + q.Duration + `" .`)

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
	server.ReadBody(r, &q)
	q.Uid = qid

	sr := server.Response{}
	_, err := time.ParseDuration(q.Duration)
	if err != nil {
		sr.Message = "Couldn't parse duration"
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
	}

	// TODO - Validate candidate fields shouldn't be empty.
	m := edit(q)
	mr := dgraph.SendMutation(m)
	if mr.Code != "ErrorOk" {
		sr.Error = mr.Message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.MarshalResponse(sr))
		return
	}

	sr.Success = true
	sr.Message = "Quiz info updated successfully."
	w.Write(server.MarshalResponse(sr))
}
