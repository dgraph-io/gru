package candidate

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/gorilla/mux"
)

type option struct {
	Id   string `json:"_uid_"`
	Name string `json:"name"`
}

type uid struct {
	Id string `json:"_uid_"`
}

type que struct {
	Uid      string  `json:"_uid_"`
	Multiple bool    `json:"multiple,string"`
	Negative float64 `json:"negative,string"`
	Positive float64 `json:"positive,string"`
	Text     string
	Name     string
	Tags     []option `json:"question.tag"`
	Options  []option `json:"question.option"`
	Correct  []option `json:"question.correct"`
}

type questions []cq

func (q questions) Len() int {
	return len(q)
}

func (q questions) Less(i, j int) bool {
	return q[i].Asked.Before(q[j].Asked)
}

func (q questions) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

type cq struct {
	Answer   string    `json:"candidate.answer"`
	Score    float64   `json:"candidate.score,string"`
	Asked    time.Time `json:"question.asked,string"`
	Answered time.Time `json:"question.answered,string"`
	Question []que     `json:"question.uid"`
}

type candidates struct {
	Id          string `json:"_uid_"`
	Name        string
	Email       string
	CandidateQn []cq `json:"candidate.question"`
	Complete    bool `json:"complete,string"`
}

type report struct {
	Candidates []candidates `json:"candidate"`
}

func reportQuery(id string) string {
	return `query {
                candidate(_uid_:` + id + `) {
                        _uid_
                        name
                        email
                        complete
                        candidate.question {
                                question.uid {
                                        _uid_
                                        text
                                        name
                                        positive
                                        negative
                                        question.tag {
                                                _uid_
                                                name
                                        }
                                        question.option {
                                                _uid_
                                                name
                                        }
                                        question.correct {
                                                _uid_
                                        }
                                multiple
                        }
                        question.asked
                        question.answered
                        candidate.answer
                        candidate.score
                }
        }
}`
}

type question struct {
	Name      string   `json:"name"`
	Multiple  bool     `json:"multiple"`
	Text      string   `json:"text"`
	TimeTaken string   `json:"time_taken"`
	Score     float64  `json:"score"`
	Options   []option `json:"options"`
	Correct   []string `json:"correct"`
	Answers   []string `json:"answers"`
	Tags      []string `json:"tags"`
}

type Summary struct {
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	TimeTaken  string     `json:"time_taken"`
	TotalScore float64    `json:"total_score"`
	Questions  []question `json:"questions"`
}

func uids(opts []option) []string {
	var ids []string
	for _, opt := range opts {
		ids = append(ids, opt.Id)
	}
	return ids
}

func names(opts []option) []string {
	var n []string
	for _, opt := range opts {
		n = append(n, opt.Name)
	}
	return n
}

func Report(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["id"]
	q := reportQuery(cid)
	b := dgraph.Query(q)

	var rep report
	sr := server.Response{}
	err := json.Unmarshal(b, &rep)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	if rep.Candidates[0].Id == "" {
		sr.Write(w, "", "Candidate not found.", http.StatusBadRequest)
		return
	}
	if !rep.Candidates[0].Complete {
		sr.Write(w, "", "Candidate hasn't completed the test.", http.StatusBadRequest)
		return
	}

	c := rep.Candidates[0]
	// TODO - Check how to obtain sorted results from Dgraph.
	sort.Sort(questions(c.CandidateQn))
	s := Summary{
		Name:  c.Name,
		Email: c.Email,
	}
	if !c.CandidateQn[len(c.CandidateQn)-1].Answered.IsZero() {
		s.TimeTaken = c.CandidateQn[len(c.CandidateQn)-1].Answered.Sub(
			c.CandidateQn[0].Asked).String()
	} else {
		// Incase we didn't record the answered for the last qn, say his
		// browser crashed.
		s.TimeTaken = c.CandidateQn[len(c.CandidateQn)-1].Asked.Sub(
			c.CandidateQn[0].Asked).String()
	}

	for _, qn := range c.CandidateQn {
		s.TotalScore += qn.Score
		q := qn.Question[0]
		sq := question{
			Name:     q.Name,
			Text:     q.Text,
			Options:  q.Options,
			Score:    qn.Score,
			Multiple: q.Multiple,
			Correct:  uids(q.Correct),
			Tags:     names(q.Tags),
			Answers:  strings.Split(qn.Answer, ","),
		}
		if qn.Answered.IsZero() {
			sq.TimeTaken = "-1"
		} else {
			sq.TimeTaken = qn.Answered.Sub(qn.Asked).String()
		}
		s.Questions = append(s.Questions, sq)
	}

	b, err = json.Marshal(s)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
