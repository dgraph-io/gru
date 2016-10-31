package report

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
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

type quiz struct {
	Duration string `json:"duration"`
}

type candidates struct {
	Id          string `json:"_uid_"`
	Name        string
	Email       string
	Feedback    string
	CandidateQn []cq   `json:"candidate.question"`
	Complete    bool   `json:"complete,string"`
	Quiz        []quiz `json:"candidate.quiz"`
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
                        feedback
                        complete
                        candidate.quiz {
                                duration
                        }
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
	Answered  bool
	Tags      []string `json:"tags"`
}

type Summary struct {
	Id         string
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Feedback   string     `json:"feedback"`
	TimeTaken  string     `json:"time_taken"`
	TotalScore float64    `json:"total_score"`
	MaxScore   float64    `json:"max_score"`
	Questions  []question `json:"questions"`
	Ip         string
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

type ReportError struct {
	Err  string
	Msg  string
	code int
}

func ReportSummary(cid string) (Summary, ReportError) {
	s := Summary{}
	s.Ip = *mail.Ip
	q := reportQuery(cid)
	var rep report
	if err := dgraph.QueryAndUnmarshal(q, &rep); err != nil {
		return s, ReportError{err.Error(), "", http.StatusInternalServerError}
	}

	if len(rep.Candidates) != 1 || rep.Candidates[0].Id == "" || len(rep.Candidates[0].Quiz) != 1 {
		return s, ReportError{"", "Candidate not found.", http.StatusBadRequest}
	}
	s.Id = rep.Candidates[0].Id

	c := rep.Candidates[0]
	s.Name = c.Name
	s.Email = c.Email
	s.Feedback = c.Feedback
	// TODO - Check how to obtain sorted results from Dgraph.
	if len(c.CandidateQn) == 0 {
		return s, ReportError{"", "Candidate hasn't started the test", http.StatusBadRequest}
	}

	sort.Sort(questions(c.CandidateQn))
	if !c.CandidateQn[len(c.CandidateQn)-1].Answered.IsZero() {
		s.TimeTaken = c.CandidateQn[len(c.CandidateQn)-1].Answered.Sub(
			c.CandidateQn[0].Asked).String()
	} else {
		// Incase we didn't record the answered for the last qn, say his
		// browser crashed or he didn't finish answering it.
		dur := c.Quiz[0].Duration
		// TODO - This is a hack because duration is stored as 0h50m0s,
		// Ideally it should be 50m0s.
		d, err := time.ParseDuration(dur)
		if err != nil {
			return s, ReportError{"", "Can't parse quiz duration.",
				http.StatusInternalServerError}
		}
		s.TimeTaken = d.String()
	}

	for _, qn := range c.CandidateQn {
		s.TotalScore += qn.Score
		q := qn.Question[0]
		s.MaxScore += float64(len(q.Correct)) * q.Positive
		answers := strings.Split(qn.Answer, ",")
		sq := question{
			Name:     q.Name,
			Text:     q.Text,
			Options:  q.Options,
			Score:    qn.Score,
			Multiple: q.Multiple,
			Correct:  uids(q.Correct),
			Tags:     names(q.Tags),
			Answers:  answers,
			Answered: len(answers) > 0 && answers[0] != "skip",
		}
		if qn.Answered.IsZero() {
			sq.TimeTaken = "-1"
		} else {
			sq.TimeTaken = qn.Answered.Sub(qn.Asked).String()
		}
		s.Questions = append(s.Questions, sq)
	}
	s.TotalScore = x.Truncate(s.TotalScore)
	s.MaxScore = x.Truncate(s.MaxScore)
	return s, ReportError{}
}

func Report(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	vars := mux.Vars(r)
	cid := vars["id"]
	s, re := ReportSummary(cid)
	if re.Msg != "" || re.Err != "" {
		sr.Write(w, re.Err, re.Msg, re.code)
		return
	}

	b, err := json.Marshal(s)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
