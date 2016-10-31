package quiz

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/admin/report"
	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// Map of candidate uids to their quiz info which is stored in Candidate
	// struct.
	cmap     map[string]Candidate
	mu       sync.RWMutex
	throttle chan time.Time
)

const (
	rate       = time.Second
	timeLayout = "2006-01-02T15:04:05Z07:00"
)

func init() {
	throttle = make(chan time.Time, 3)
	go rateLimit()
	cmap = make(map[string]Candidate)
}

type Answer struct {
	Id   string `json:"_uid_"`
	Text string `json:"name"`
}

type Question struct {
	Id string `json:"_uid_"`

	// cuid represents the uid of the question asked to the candidate, it is linked
	// to the original question _uid_.
	Cid     string   `json:"cuid"`
	Text    string   `json:"text"`
	Options []Answer `json:"question.option"`
	// TODO - Remove the ,string after we incorporate Dgraph schema here.
	IsMultiple bool    `json:"multiple,string"`
	Positive   float64 `json:"positive,string"`
	Negative   float64 `json:"negative,string"`
	// Score of the candidate is sent as part of the questions API.
	Score float64 `json:"score"`
}

// Candidate is used to keep track of the state of the quiz for a candidate.
type Candidate struct {
	name         string
	token        string
	score        float64
	qns          []Question
	lastExchange time.Time
	quizDuration time.Duration
	quizStart    time.Time
	validity     time.Time

	// We use these so that we can show candidate the same question if he
	// refreshes the page/recovers from a crash.
	lastQnUid  string
	lastQnCuid string
}

func updateMap(uid string, c Candidate) {
	mu.Lock()
	defer mu.Unlock()
	cmap[uid] = c
}

func readMap(uid string) (Candidate, error) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := cmap[uid]
	if !ok {
		return Candidate{}, fmt.Errorf("Uid not found in map.")
	}
	return c, nil
}

type quiz struct {
	Id        string     `json:"_uid_"`
	Duration  string     `json:"duration"`
	Questions []Question `json:"quiz.question"`
}

type quizInfo struct {
	Quizzes []quiz `json:"quiz"`
}

func quizQns(quizId string, qnsAsked []string) ([]Question, error) {
	q := `{
		quiz(_uid_: ` + quizId + `) {
			quiz.question {
				_uid_
				text
				positive
				negative
				question.option {
					_uid_
					name
				}
				multiple
			}
		}
	}`

	var resp quizInfo
	if err := dgraph.QueryAndUnmarshal(q, &resp); err != nil {
		return []Question{}, err
	}
	if len(resp.Quizzes) != 1 {
		return []Question{}, fmt.Errorf("Expected length of quizzes: %v. Got %v",
			1, len(resp.Quizzes))
	}

	if len(qnsAsked) == 0 {
		return resp.Quizzes[0].Questions, nil
	}

	allQns := resp.Quizzes[0].Questions
	idx := 0
	for _, qn := range allQns {
		if x.StringInSlice(qn.Id, qnsAsked) == -1 {
			allQns[idx] = qn
			idx++
		}
	}
	allQns = allQns[:idx]
	return allQns, nil
}

// Used to fetch data about a candidate from Dgraph and populate Candidate struct.
type cand struct {
	Name       string
	Token      string    `json:"token"`
	Validity   string    `json:"validity"`
	Complete   bool      `json:"complete,string"`
	Quiz       []quiz    `json:"candidate.quiz"`
	Questions  []qids    `json:"candidate.question"`
	QuizStart  time.Time `json:"quiz_start"`
	LastQnUid  string    `json:"candidate.lastqnuid"`
	LastQnCuid string    `json:"candidate.lastqncuid"`
}

type resp struct {
	Cand []cand `json:"quiz.candidate"`
}

type uid struct {
	Id string `json:"_uid_"`
}

type qids struct {
	QuestionUid []uid   `json:"question.uid"`
	Score       float64 `json:"candidate.score,string"`
	Answered    string  `json:"question.answered"`
}

func timeLeft(start time.Time, dur time.Duration) time.Duration {
	if start.IsZero() {
		return dur
	}
	// If start isn't zero we return the time left.
	return start.Add(dur).Sub(time.Now())
}

// Checks the JWT Token and gets the user id from the claims.
func validateToken(r *http.Request) (string, error) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Bearer" {
		return "", fmt.Errorf("Format of authorization header isn't correct")
	}
	token, err := jwt.ParseWithClaims(s[1], &x.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(*auth.Secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*x.Claims); ok && claims.UserId != "" {
		return claims.UserId, nil
	}
	return "", fmt.Errorf("Invalid JWT token")
}

func sendReport(cid string) {
	dir, _ := os.Getwd()
	t, err := template.ParseFiles(filepath.Join(dir, "quiz/report.html"))
	if err != nil {
		fmt.Println(err)
	}
	buf := new(bytes.Buffer)
	s, re := report.ReportSummary(cid)
	if re.Err != "" || re.Msg != "" {
		fmt.Printf("Error: %v with msg: %v while generating report.",
			re.Err, re.Msg)
		return
	}
	if err = t.Execute(buf, s); err != nil {
		fmt.Println(err)
	}
	mail.SendReport(s.Name, s.TotalScore, s.MaxScore, buf.String())
}

func candQuery(cid string) string {
	return `{
        quiz.candidate(_uid_:` + cid + `) {
                name
                email
                token
                validity
                complete
                quiz_start
                candidate.quiz {
                        _uid_
                        duration
                }
                candidate.question {
                        question.uid {
                                _uid_
                        }
                        question.answered
                        candidate.score
                }
                candidate.lastqnuid
                candidate.lastqncuid
          }
    }`
}

// Checks for candidate in cache, if we find it then we return. Else we load up
// information from the Database into the cache.
func checkAndUpdate(uid string) (int, error) {
	if _, err := readMap(uid); err == nil {
		// Got candidate information in Cache, return.
		return http.StatusOK, nil
	}

	// Candidate doesn't exist in the map. So we get candidate info from database
	// and insert it into map.
	q := candQuery(uid)
	var resp resp
	if err := dgraph.QueryAndUnmarshal(q, &resp); err != nil {
		return http.StatusInternalServerError, err
	}

	if len(resp.Cand) != 1 || len(resp.Cand[0].Quiz) != 1 {
		// No candidiate found with given uid
		return http.StatusUnauthorized, fmt.Errorf("Candidate not found")
	}

	cand := resp.Cand[0]
	quiz := cand.Quiz[0]
	if cand.Complete {
		return http.StatusUnauthorized, fmt.Errorf("You have already completed the quiz.")

	}
	if quiz.Id == "" {
		return http.StatusUnauthorized, fmt.Errorf("Invalid token.")

	}

	c := Candidate{
		quizStart:  cand.QuizStart,
		lastQnUid:  cand.LastQnUid,
		lastQnCuid: cand.LastQnCuid,
		name:       cand.Name,
		token:      cand.Token,
	}
	// TODO - Check how can we store this in appropriate format so that explicit parsing isn't
	// required.
	var err error
	if c.validity, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", cand.Validity); err != nil {
		return http.StatusInternalServerError, err
	}
	if c.validity.Before(time.Now()) {
		return http.StatusUnauthorized,
			fmt.Errorf("Your token has already expired. Please mail us at contact@dgraph.io.")
	}

	// We check that quiz duration hasn't elapsed in case the candidate tries
	// to validate again say after a browser crash.
	if c.quizDuration, err = time.ParseDuration(quiz.Duration); err != nil {
		return http.StatusInternalServerError, err
	}

	if timeLeft(c.quizStart, c.quizDuration) < 0 {
		return http.StatusUnauthorized, fmt.Errorf("Your token is no longer valid.")
	}

	var qa []string
	if len(cand.Questions) > 0 {
		// He has already been asked some questions. Lets figure out the
		// ones he has answered.
		qa = qnsAnswered(cand.Questions)
	}

	// Get quiz questions for the quiz id.
	qnsUnanswered, err := quizQns(quiz.Id, qa)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	shuffleQuestions(qnsUnanswered)
	// Lets bring the last question asked to the first place.
	for idx, qn := range qnsUnanswered {
		if qn.Id == cand.LastQnUid {
			qnsUnanswered[0], qnsUnanswered[idx] = qnsUnanswered[idx], qnsUnanswered[0]
			break
		}
	}
	c.qns = qnsUnanswered
	c.score = calcScore(cand.Questions)
	updateMap(uid, c)
	return http.StatusOK, nil
}
