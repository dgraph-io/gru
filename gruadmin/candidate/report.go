package candidate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/gorilla/mux"
)

type option struct {
	Id   string `json:"_uid_"`
	Name string `json:"name"`
}

type que struct {
	Uid      string   `json:"_uid_"`
	Multiple bool     `json:"multiple,string"`
	Negative float64  `json:"negative,string"`
	Positive float64  `json:"positive,string"`
	Text     string   `json:"text"`
	Tags     []option `json:"question.tag"`
	Options  []option `json:"question.option"`
	Correct  []option `json:"question.correct"`
}

type cq struct {
	Answer   string    `json:"candidate.answer"`
	Score    float64   `json:"candidate.score,string"`
	Asked    time.Time `json:"question.asked,string"`
	Answered time.Time `json:"question.answered,string"`
	Question []que     `json:"question.uid"`
}

type candidates struct {
	Id          string `json:"_uid"`
	CandidateQn []cq   `json:"candidate.question"`
	Complete    bool   `json:"complete,string"`
}

type report struct {
	Candidates []candidates `json:"candidate"`
}

func reportQuery(id string) string {
	return `query {
        candidate(_uid_:` + id + `) {
			    _uid_
				complete
				candidate.question {
                        question.uid {
                                _uid_
                                text
                                positive
                                negative
                                question.tag {
									_uid_
									name
                                }
                                question.option {
                                        _uid_
                                        name
                                }
                                question.correct {
                                        _uid_
                                        name
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

func Report(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["id"]
	q := reportQuery(cid)
	b := dgraph.Query(q)
	fmt.Println(string(b))

	var rep report
	sr := server.Response{}
	err := json.Unmarshal(b, &rep)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", rep)
	if rep.Candidates[0].Id == "" {
		sr.Write(w, "", "Candidate not found.", http.StatusBadRequest)
		return
	}
	if !rep.Candidates[0].Complete {
		sr.Write(w, "", "Candidate hasn't completed the test.", http.StatusBadRequest)
		return
	}
}
