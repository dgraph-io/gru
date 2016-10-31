package quiz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/admin/report"
	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
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
	score        float64
	qns          []Question
	lastExchange time.Time
	quizDuration time.Duration
	quizStart    time.Time
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

type qnIdsResp struct {
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
	res, err := dgraph.Query(q)
	if err != nil {
		return []Question{}, err
	}
	var resp qnIdsResp
	json.Unmarshal(res, &resp)
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
		if !x.StringInSlice(qn.Id, qnsAsked) {
			allQns[idx] = qn
			idx++
		}
	}
	allQns = allQns[:idx]
	return allQns, nil
}

// Used to fetch data about a candidate from Dgraph and populate Candidate struct.
type cand struct {
	Name      string
	Token     string    `json:"token"`
	Validity  string    `json:"validity"`
	Complete  bool      `json:"complete,string"`
	Quiz      []quiz    `json:"candidate.quiz"`
	Questions []qids    `json:"candidate.question"`
	QuizStart time.Time `json:"quiz_start"`
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
}

func validate(cid string) string {
	return `{
	quiz.candidate(_uid_:` + cid + `) {
		name
		email
		token
		validity
		complete
		candidate.quiz {
			_uid_
			duration
		}
		candidate.question {
			question.uid {
				_uid_
			}
		}
	  }
    }`
}

func timeLeft(start time.Time, dur time.Duration) time.Duration {
	if start.IsZero() {
		return dur
	}
	// If start isn't zero we return the time left.
	return start.Add(dur).Sub(time.Now())
}

func rateLimit() {
	rateTicker := time.NewTicker(rate)
	defer rateTicker.Stop()

	for t := range rateTicker.C {
		select {
		case throttle <- t:
		default:
		}
	}
}

type validateRes struct {
	Token    string `json:"token"`
	Duration string `json:"duration"`
	Started  bool   `json:"quiz_started"`
	Name     string
}

func Validate(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	select {
	case <-throttle:
		break
	case <-time.After(rate):
		sr.Write(w, "", "Too many requests. Please try after again.",
			http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	// This is the length of the random string. The id is uid + random string.
	if len(id) < 33 {
		sr.Write(w, "", "Invalid token.", http.StatusUnauthorized)
		return
	}

	uid, token := id[:len(id)-33], id[len(id)-33:]

	c, err := readMap(uid)
	// Check for duplicate session.
	if err == nil && !c.lastExchange.IsZero() {
		timeSinceLastExchange := time.Now().Sub(c.lastExchange)
		// To avoid duplicate sessions.
		if timeSinceLastExchange < 10*time.Second {
			sr.Write(w, "", "You have another active session. Please try after some time.",
				http.StatusUnauthorized)
			return
		}
	}

	// Candidate doesn't exist in the map. So we get candidate info from uid and
	// insert it into map.
	q := validate(uid)
	res, err := dgraph.Query(q)
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	var resp resp
	json.Unmarshal(res, &resp)
	if len(resp.Cand) != 1 || len(resp.Cand[0].Quiz) != 1 {
		// No candidiate found with given uid
		sr.Write(w, "", "Candidate not found", http.StatusUnauthorized)
		return
	}
	if resp.Cand[0].Complete {
		sr.Write(w, "", "You have already completed the quiz.", http.StatusUnauthorized)
		return
	}
	if resp.Cand[0].Token != token || resp.Cand[0].Quiz[0].Id == "" {
		sr.Write(w, "", "Invalid token.", http.StatusUnauthorized)
		return
	}

	var v time.Time
	if v, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", resp.Cand[0].Validity); err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	if v.Before(time.Now()) {
		sr.Write(w, "", "Your token has already expired. Please contact contact@dgraph.io.",
			http.StatusUnauthorized)
		return
	}

	cand := resp.Cand[0]
	quiz := cand.Quiz[0]
	dur, err := time.ParseDuration(quiz.Duration)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	if timeLeft(c.quizStart, dur) < 0 {
		sr.Write(w, "", "Your token is no longer valid.", http.StatusUnauthorized)
		return
	}
	// He has already been asked some questions.
	var qa []string
	if len(cand.Questions) > 0 {
		qa = qnsAsked(cand.Questions)
	}

	// Get quiz questions for the quiz id.
	qns, err := quizQns(quiz.Id, qa)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
	}
	shuffleQuestions(qns)

	if len(cand.Questions) > 0 {
		c.qns = qns
		updateMap(uid, c)
	} else {
		c := Candidate{
			qns:          qns,
			quizDuration: dur,
		}
		updateMap(uid, c)
	}

	// Add the user id as a claim and return the token.
	claims := x.Claims{
		UserId: uid,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenString, err := jwtToken.SignedString([]byte(*auth.Secret))
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(validateRes{
		Token:    tokenString,
		Duration: timeLeft(c.quizStart, dur).String(),
		// Whether quiz was already started by the candidate.
		// If this is true the client can just call the questions API and
		// skip showing the instructions page.
		Started: len(cand.Questions) > 0,
		Name:    cand.Name,
	})
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

func QuestionHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validateToken(r); err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
		return
	}

	var c Candidate
	if c, err = checkCand(userId); err != nil {
		sr.Write(w, err.Error(), "", http.StatusBadRequest)
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
			  <_uid_:` + userId + `> <quiz_start> "` + c.quizStart.Format(timeLayout) + `" .
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
	updateMap(userId, c)
	// Truncate score to two decimal places.
	qn.Score = x.Truncate(c.score)
	qn.Cid = res.Uids["qn"]
	b, err := json.Marshal(qn)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Write(b)
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

type qa struct {
	Answered string `json:"question.answered"`
}

type checkAnswer struct {
	Question []qa `json:"candidate.question"`
}

// Queries Dgraph and checks if the candidate has already answered the question.
func checkAnswered(cuid string) (int, error) {
	q := `{
		candidate.question(_uid_:` + cuid + `) {
			question.answered
		}
	}`
	b, err := dgraph.Query(q)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var ca checkAnswer
	err = json.Unmarshal(b, &ca)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if len(ca.Question) != 1 || ca.Question[0].Answered != "" {
		return http.StatusBadRequest, fmt.Errorf("You have already answered this question.")

	}
	return http.StatusOK, nil
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validateToken(r); err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
		return
	}

	var c Candidate
	if c, err = checkCand(userId); err != nil {
		sr.Write(w, err.Error(), "", http.StatusBadRequest)
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
	answerIds := strings.Split(aid, ",")
	if cuid == "" || len(answerIds) == 0 {
		sr.Write(w, "Answer ids/cuid can't be empty", "", http.StatusBadRequest)
		return
	}

	if status, err := checkAnswered(cuid); err != nil {
		sr.Write(w, err.Error(), "", status)
		return
	}

	m, err := qnMeta(qid)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
	}
	score := isCorrectAnswer(answerIds, m.correct, m.positive, m.negative)
	c.score = c.score + score
	updateMap(userId, c)
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

// checks for candidate in the map, if not present it checks the
// database and loads his info. This would help recover from server
// crashes.
func checkCand(uid string) (Candidate, error) {
	c, err := readMap(uid)
	if err == nil {
		return c, nil
	}
	c, err = Load(uid)
	if err != nil {
		return c, err
	}
	return c, nil
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validateToken(r); err != nil {
		sr.Write(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	var c Candidate
	if c, err = checkCand(userId); err != nil {
		sr.Write(w, err.Error(), "", http.StatusBadRequest)
		return
	}
	c.lastExchange = time.Now()
	updateMap(userId, c)
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

func Feedback(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validateToken(r); err != nil {
		sr.Write(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	feedback := r.PostFormValue("feedback")
	if feedback == "" {
		sr.Write(w, "", "Feedback can't be empty", http.StatusBadRequest)
		return
	}
	m := `	mutation {
			set {
				<_uid_:` + userId + `> <feedback> "` + feedback + `" .
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
	return
}

func load(cid string) string {
	return `{
	quiz.candidate(_uid_:` + cid + `) {
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
			candidate.score
		}
	  }
    }`
}

// TODO - Make code dry abstract out logic here and in validate.
// That will also fix bug where validate doesn't update quizStart from DB
// after server restarts.
func Load(uid string) (Candidate, error) {
	c := Candidate{}
	q := load(uid)
	res, err := dgraph.Query(q)
	if err != nil {
		return c, err
	}

	var resp resp
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return c, err
	}

	if len(resp.Cand) != 1 || len(resp.Cand[0].Quiz) != 1 {
		// No candidiate found with given uid
		return c, fmt.Errorf("Candidate not found.")
	}
	if resp.Cand[0].Complete {
		return c, fmt.Errorf("Already completed the quiz.")
	}
	// Means we found a candidate who has not completed the quiz.
	cand := resp.Cand[0]
	quiz := resp.Cand[0].Quiz[0]
	c.quizDuration, err = time.ParseDuration(quiz.Duration)
	if err != nil {
		return c, err
	}
	c.quizStart = cand.QuizStart

	var qa []string
	if len(cand.Questions) > 0 {
		qa = qnsAsked(cand.Questions)
		fmt.Println("len qnsAsked", len(qa))
	}

	// Get quiz questions for the quiz id.
	qns, err := quizQns(quiz.Id, qa)
	if err != nil {
		return c, err
	}

	shuffleQuestions(qns)
	c.qns = qns
	c.score = calcScore(cand.Questions)
	updateMap(uid, c)
	return c, nil
}
