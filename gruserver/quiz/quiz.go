package quiz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/mail"
	"github.com/dgraph-io/gru/gruadmin/report"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
)

type Answer struct {
	Id   string `json:"_uid_"`
	Text string `json:"name"`
}

type Question struct {
	Id string `json:"_uid_"`

	// cuid represents the uid of thequestion asked to the candidate, it is linked
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

type Candidate struct {
	score float64
	qns   []Question
	// Used to check the order of answers.
	lastQnId     string
	lastExchange time.Time
	quizDuration time.Duration
	quizStart    time.Time
}

func (c *Candidate) QuizStart() time.Time {
	return c.quizStart
}

func (c Candidate) LastExchange() time.Time {
	return c.lastExchange
}

func New(uid string, qns []Question, qd time.Duration) Candidate {
	c := Candidate{}
	c.qns = make([]Question, len(qns))
	c.quizDuration = qd
	copy(c.qns, qns)
	UpdateMap(uid, c)
	return c
}

func Update(uid string, qns []Question) {
	c, err := ReadMap(uid)
	if err != nil {
		return
	}
	c.qns = qns
	UpdateMap(uid, c)
}

func init() {
	// TODO - Handler server crashes and restarts. That would mean reload cmap from DB.
	cmap = make(map[string]Candidate)
}

var (
	cmap map[string]Candidate
	mu   sync.RWMutex
)

func UpdateMap(uid string, c Candidate) {
	mu.Lock()
	defer mu.Unlock()
	cmap[uid] = c
}

func ReadMap(uid string) (Candidate, error) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := cmap[uid]
	if !ok {
		return Candidate{}, fmt.Errorf("Uid not found in map.")
	}
	return c, nil
}

func validate(r *http.Request) (string, error) {
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

	if claims, ok := token.Claims.(*x.Claims); ok {
		return claims.UserId, nil
	}
	return "", fmt.Errorf("Cannot parse claims.")
}

func sendReport(cid string) {
	t, err := template.ParseFiles("../gruserver/quiz/report.html")
	if err != nil {
		fmt.Println(err)
	}
	buf := new(bytes.Buffer)
	s, re := report.ReportSummary(cid)
	if re.Err != "" || re.Msg != "" {
		return
	}
	if err = t.Execute(buf, s); err != nil {
		fmt.Println(err)
	}
	body := buf.String()
	mail.SendReport(s.Name, s.TotalScore, s.MaxScore, body)
}

func shuffleOptions(opts []Answer) {
	for i := range opts {
		j := rand.Intn(i + 1)
		opts[i], opts[j] = opts[j], opts[i]
	}
}

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validate(r); err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		sr.Write(w, err.Error(), "User not found.", http.StatusBadRequest)
		return
	}

	if !c.quizStart.IsZero() && time.Now().After(c.quizStart.Add(c.quizDuration)) {
		sr.Write(w, "", "Your quiz has already finished.",
			http.StatusBadRequest)
		return
	}

	// This means its the first question he is being asked.
	// If this is because the server crashed then we should have recovered before
	// the candidate reaches here.
	if c.quizStart.IsZero() {
		c.quizStart = time.Now().UTC()
		m := `mutation {
		  set {
			  <_uid_:` + userId + `> <quiz_start> "` + c.quizStart.String() + `" .
			}
		}
		`
		res, err := dgraph.SendMutation(m)
		if err != nil {
			sr.Write(w, "", err.Error(), http.StatusInternalServerError)
			return
		}
		if res.Code != "ErrorOk" {
			sr.Write(w, res.Message, "", http.StatusInternalServerError)
			return
		}
	}

	if len(c.qns) == 0 {
		q := Question{
			Id:    "END",
			Score: float64(int(c.score*100)) / 100,
		}
		m := `mutation {
		  set {
			  <_uid_:` + userId + `> <complete> "true" .
			}
		}
		`
		res, err := dgraph.SendMutation(m)
		if err != nil {
			sr.Write(w, "", err.Error(), http.StatusInternalServerError)
			return
		}
		if res.Code != "ErrorOk" {
			sr.Write(w, res.Message, "", http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(q)
		if err != nil {
			sr.Write(w, err.Error(), "", http.StatusInternalServerError)
			return
		}
		go sendReport(userId)
		w.Write(b)
		return
	}

	qn := c.qns[0]
	shuffleOptions(qn.Options)
	m := `mutation {
		set {
			<_uid_:` + userId + `> <candidate.question> <_new_:qn> .
			<_new_:qn> <question.uid> <_uid_:` + qn.Id + `> .
			<_uid_:` + qn.Id + `> <question.candidate> <_uid_:` + userId + `> .
			<_new_:qn> <question.asked> "` + time.Now().Format("2006-01-02T15:04:05Z07:00") + `" .
		}
	}`

	res, err := dgraph.SendMutation(m)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if res.Code != "ErrorOk" || res.Uids["qn"] == "" {
		sr.Write(w, res.Message, "", http.StatusInternalServerError)
		return
	}

	c.qns = c.qns[1:]
	c.lastQnId = qn.Id
	UpdateMap(userId, c)
	// Truncate score to two decimal places.
	qn.Score = float64(int(c.score*100)) / 100
	qn.Cid = res.Uids["qn"]
	b, err := json.Marshal(qn)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func isCorrectAnswer(selected []string, actual []string, pos, neg float64) float64 {
	if selected[0] == "skip" {
		return 0
	}
	// For multiple choice qnstions, we have partial scoring.
	if len(actual) == 1 {
		if selected[0] == actual[0] {
			return pos
		}
		return -neg
	}
	var score float64
	for _, aid := range selected {
		correct := false
		for _, caid := range actual {
			if caid == aid {
				correct = true
				break
			}
		}
		if correct {
			score += pos
		} else {
			score -= neg
		}
	}
	return score
}

type correct struct {
	Uid string `json:"_uid_"`
}

// Used to marshal response from Dgraph.
type questionMeta struct {
	Negative float64 `json:"negative,string"`
	Positive float64 `json:"positive,string"`
	// TODO - Maybe store correct later as a comma separated string uids so that
	// processing isn't required.
	Correct []correct `json:"question.correct"`
}

type qmRes struct {
	QuestionMeta []questionMeta `json:"question"`
}

type questionCorrectMeta struct {
	negative float64
	positive float64
	correct  []string
}

func qnMeta(qid string) (questionCorrectMeta, error) {
	q := `{
        question(_uid_: ` + qid + `) {
                question.correct {
                _uid_
        }
        positive
        negative
        }
}`
	res, err := dgraph.Query(q)
	if err != nil {
		return questionCorrectMeta{}, err
	}
	var resp qmRes
	json.Unmarshal(res, &resp)

	if len(resp.QuestionMeta) != 1 {
		return questionCorrectMeta{},
			fmt.Errorf("There should be just one question returned")
	}
	question := resp.QuestionMeta[0]
	// TODO - Maybe cache this stuff later.
	correctAnswers := []string{}
	for _, answer := range question.Correct {
		correctAnswers = append(correctAnswers, answer.Uid)
	}

	return questionCorrectMeta{
		negative: question.Negative,
		positive: question.Positive,
		correct:  correctAnswers,
	}, nil
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validate(r); err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusBadRequest)
		return
	}

	if !c.quizStart.IsZero() && time.Now().After(c.quizStart.Add(c.quizDuration)) {
		sr.Write(w, "", "Your quiz has already finished.",
			http.StatusBadRequest)
		return
	}

	qid := r.PostFormValue("qid")
	aid := r.PostFormValue("aid")
	cuid := r.PostFormValue("cuid")
	if qid != c.lastQnId {
		sr.Write(w, "Questions can only be answered in sequential order",
			"", http.StatusBadRequest)
		return
	}

	answerIds := strings.Split(aid, ",")
	if cuid == "" || len(answerIds) == 0 {
		sr.Write(w, "Answer ids/cuid can't be empty",
			"", http.StatusBadRequest)
		return
	}

	m, err := qnMeta(qid)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
	}
	score := isCorrectAnswer(answerIds, m.correct, m.positive, m.negative)
	c.score = c.score + score
	UpdateMap(userId, c)
	mutation := `mutation {
		set {
			<_uid_:` + cuid + `> <candidate.answer> "` + aid + `" .
			<_uid_:` + cuid + `> <candidate.score> "` + strconv.FormatFloat(score, 'g', -1, 64) + `" .
			<_uid_:` + cuid + `> <question.answered> "` + time.Now().Format("2006-01-02T15:04:05Z07:00") + `" .
		}
	}`
	res, err := dgraph.SendMutation(mutation)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	if res.Code != "ErrorOk" {
		sr.Write(w, res.Message, "", http.StatusInternalServerError)
		return
	}
}

type pingRes struct {
	TimeLeft string `json:"time_left"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validate(r); err != nil {
		sr.Write(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	c, err := ReadMap(userId)
	if err != nil {
		sr.Write(w, err.Error(), "User not found", http.StatusBadRequest)
		return
	}

	c.lastExchange = time.Now()
	UpdateMap(userId, c)
	pr := &pingRes{TimeLeft: "-1"}
	if !c.quizStart.IsZero() {
		end := c.quizStart.Add(c.quizDuration).Truncate(time.Second)
		timeLeft := end.Sub(time.Now().UTC().Truncate(time.Second))
		if timeLeft <= 0 {
			m := `mutation {
			set {
				<_uid_:` + userId + `> <complete> "true" .
			}
			}
			`
			res, err := dgraph.SendMutation(m)
			if err != nil {
				sr.Write(w, "", err.Error(), http.StatusInternalServerError)
				return
			}
			if res.Code != "ErrorOk" {
				sr.Write(w, res.Message, "", http.StatusInternalServerError)
				return
			}
			go sendReport(userId)
		}
		pr.TimeLeft = timeLeft.String()
	}
	json.NewEncoder(w).Encode(pr)
}
