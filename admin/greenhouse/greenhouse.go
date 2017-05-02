package greenhouse

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/gru/admin/candidate"
	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
)

const (
	baseUrl    = "https://gru.dgraph.io"
	dateFormat = "Jan 2 15:04:05 2006"
	week       = 7 * 24 * time.Hour
)

var (
	ghKey = flag.String("gh", "", "Api key sent as username in basic auth by Greenhouse")
)

type outgoingQuiz struct {
	Id   string `json:"partner_test_id"`
	Name string `json:"partner_test_name"`
}

type quiz struct {
	Id   string `json:"_uid_"`
	Name string `json:"name"`
}

type meValue struct {
	Quizzes []quiz `json:"quiz"`
}

type allTests struct {
	Me []meValue `json:"me"`
}

func validAuth(r *http.Request) bool {
	username, _, ok := r.BasicAuth()
	if !ok {
		return false
	}

	if username != *ghKey {
		return false
	}
	return true
}

// TODO - Delete name test from Gru prod
func Tests(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	if valid := validAuth(r); !valid {
		sr.Write(w, "Authorization header is incorrect", "", http.StatusUnauthorized)
		return
	}

	q := `{
    	  me(id: root) {
            quiz {
	      _uid_
 	      name
    	    }
    	  }
        }`

	var at allTests
	err := dgraph.QueryAndUnmarshal(q, &at)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if len(at.Me) == 0 {
		sr.Write(w, "", "Expected atleast one quiz, got zero", http.StatusInternalServerError)
		return
	}
	quizzes := at.Me[0].Quizzes
	if len(quizzes) == 0 {
		sr.Write(w, "", "Expected atleast one quiz, got zero", http.StatusInternalServerError)
		return
	}

	oq := make([]outgoingQuiz, 0, len(quizzes))
	for _, q := range quizzes {
		oq = append(oq, outgoingQuiz{
			Id:   q.Id,
			Name: q.Name,
		})
	}

	server.MarshalAndWrite(w, &oq)
}

type cand struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type sendTest struct {
	QuizId    string `json:"partner_test_id"`
	Candidate cand   `json:"candidate"`
}

type sendTestRes struct {
	Id string `json:"partner_interview_id"`
}

func SendTest(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	if valid := validAuth(r); !valid {
		sr.Write(w, "Authorization header is incorrect", "", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var st sendTest
	err := decoder.Decode(&st)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusBadRequest)
		return
	}

	candQuizId, err := candidate.AddCand(st.QuizId, strings.Join([]string{st.Candidate.FirstName, st.Candidate.LastName}, " "),
		st.Candidate.Email, time.Now().Add(2*week))
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}

	var res sendTestRes
	res.Id = candQuizId
	server.MarshalAndWrite(w, &res)
}

type metadata struct {
	Feedback    string
	StartedAt   string
	CompletedAt string
}

type testStatus struct {
	Status     string   `json:"partner_status"`
	ProfileUrl string   `json:"partner_profile_url"`
	Score      float64  `json:"partner_score"`
	Metadata   metadata `json:"metadata"`
}

type candidates struct {
	Id        string `json:"_uid_"`
	Feedback  string
	Score     float64
	Complete  bool
	QuizStart time.Time `json:"quiz_start"`
	QuizEnd   time.Time `json:"completed_at"`
}

type candQuizResp struct {
	Candidates []candidates `json:"me"`
}

func reportUrl(interviewId string) string {
	return fmt.Sprintf("%s/#/admin/invite/candidate-report/%s", baseUrl, interviewId)
}

func TestStatus(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	if valid := validAuth(r); !valid {
		sr.Write(w, "Authorization header is incorrect", "", http.StatusUnauthorized)
		return
	}

	interviewId := r.URL.Query().Get("partner_interview_id")
	if interviewId == "" {
		sr.Write(w, "partner_interview_id can't be empty", "", http.StatusBadRequest)
		return
	}

	q := `{
    	  me(id: ` + interviewId + `) {
    	    _uid_
	    complete
	    score
	    quiz_start
	    completed_at
	    feedback
	  }
        }`

	var resp candQuizResp
	if err := dgraph.QueryAndUnmarshal(q, &resp); err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Candidates) == 0 {
		sr.Write(w, "No quiz found with given id, "+interviewId, "", http.StatusBadRequest)
		return
	}

	cand := resp.Candidates[0]
	var ts testStatus
	if cand.Complete == false {
		ts.Status = "incomplete"
		server.MarshalAndWrite(w, &ts)
		return
	}

	ts.Status = "complete"
	ts.Score = cand.Score
	ts.ProfileUrl = reportUrl(interviewId)
	ts.Metadata = metadata{
		Feedback:    cand.Feedback,
		StartedAt:   cand.QuizStart.Format(dateFormat),
		CompletedAt: cand.QuizEnd.Format(dateFormat),
	}
	server.MarshalAndWrite(w, &ts)
}

type requestError struct {
	ApiCall            string   `json:"api_call"`
	Errors             []string `json:"errors"`
	PartnerTestId      string   `json:"partner_test_id"`
	PartnerTestName    string   `json:"partner_test_name"`
	PartnerInterviewId string   `json:"partner_interview_id"`
	CandEmail          string   `json:"candidate_email"`
}

func RequestErrors(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	if valid := validAuth(r); !valid {
		sr.Write(w, "Authorization header is incorrect", "", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var re requestError
	err := decoder.Decode(&re)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusBadRequest)
	}

	// TODO - Integrate with sentry.
	fmt.Printf("Error: %+v\n", re)
	w.WriteHeader(http.StatusOk)
}
