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
	Uid       string `json:"uid"`
	Name      string
	Duration  int
	Cutoff    float64    `json:"cut_off"`
	Threshold float64    `json:"threshold"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Uid       string `json:"uid"`
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
	m.SetString("_:quiz", "is_quiz", "")
	// TODO - Error if Name is empty.
	m.SetString("_:quiz", "name", q.Name)
	m.SetString("_:quiz", "threshold", strconv.FormatFloat(q.Threshold, 'g', -1, 64))
	m.SetString("_:quiz", "cut_off", strconv.FormatFloat(q.Cutoff, 'g', -1, 64))
	m.SetString("_:quiz", "duration", strconv.Itoa(q.Duration))
	for _, q := range q.Questions {
		m.SetLink("_:quiz", "quiz.question", q.Uid)
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
	sr.Message = "Quiz Successfully Saved!"
	w.Write(server.MarshalResponse(sr))
}

func Index(w http.ResponseWriter, r *http.Request) {
	q := `{
		quizzes(func: has(is_quiz)) {
			uid
			name
			duration
			quiz.question {
				uid
				text
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

func getQuizQuery(quizId string) string {
	return `{
		quiz(func: uid(` + quizId + `)) {
			uid
			name
			duration
			cut_off
			threshold
			quiz.question {
				uid
				name
				text
				positive
				negative
				tags: question.tag {
					uid
					name
				}
				correct: question.correct {
					uid
					name
				}
			}
		}
  }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	res, err := dgraph.Query(getQuizQuery(vars["id"]))
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func edit(q Quiz) *dgraph.Mutation {
	m := new(dgraph.Mutation)
	// TODO - Validate these fields.
	m.SetString(q.Uid, "is_quiz", "")
	m.SetString(q.Uid, "name", q.Name)
	m.SetString(q.Uid, "duration", strconv.Itoa(q.Duration))
	m.SetString(q.Uid, "threshold", strconv.FormatFloat(q.Threshold, 'g', -1, 64))
	m.SetString(q.Uid, "cut_off", strconv.FormatFloat(q.Cutoff, 'g', -1, 64))

	// Create and associate Tags
	for _, que := range q.Questions {
		if que.Is_delete {
			m.DelLink(q.Uid, "quiz.question", que.Uid)
		} else if que.Uid != "" {
			m.SetLink(q.Uid, "quiz.question", que.Uid)
		}
	}
	return m
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
