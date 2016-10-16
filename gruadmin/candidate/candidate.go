package candidate

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/mail"
	"github.com/dgraph-io/gru/gruadmin/server"
	quizp "github.com/dgraph-io/gru/gruserver/quiz"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var throttle chan time.Time

func init() {
	throttle = make(chan time.Time, 3)
	go rateLimit()
}

type qids struct {
	QuestionUid []uid `json:"question.uid"`
}

type Candidate struct {
	Uid       string
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	Validity  string `json:"validity"`
	Complete  bool   `json:"complete,string"`
	QuizId    string `json:"quiz_id"`
	OldQuizId string `json:"old_quiz_id"`
	Quiz      []quiz `json:"candidate.quiz"`
	Questions []qids `json:"candidate.question"`
}

const (
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	validityLayout = "2006-01-02"
	rate           = time.Second
)

// TODO - Optimize later.
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func index(quizId string) string {
	return `
	{
	quiz(_uid_: ` + quizId + `) {
		quiz.candidate {
			_uid_
			name
			email
			validity
			complete
		}
	}
}
`
}

func Index(w http.ResponseWriter, r *http.Request) {
	quizId := r.URL.Query().Get("quiz_id")
	if quizId == "" {
		sr := server.Response{}
		sr.Write(w, "", "Quiz id can't be empty.", http.StatusBadRequest)
		return
	}
	q := index(quizId)
	res := dgraph.Query(q)
	w.Write(res)
}

func add(c Candidate) string {
	return `
	mutation {
		set {
		<_uid_:` + c.QuizId + `> <quiz.candidate> <_new_:c> .
		<_new_:c> <candidate.quiz> <_uid_:` + c.QuizId + `> .
		<_new_:c> <email> "` + c.Email + `" .
		<_new_:c> <name> "` + c.Name + `" .
		<_new_:c> <token> "` + c.Token + `" .
		<_new_:c> <validity> "` + c.Validity + `" .
		<_new_:c> <complete> "false" .
		}
	}`
}

func Add(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	var c Candidate
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var t time.Time
	if t, err = time.Parse(validityLayout, c.Validity); err != nil {
		sr.Message = "Couldn't parse the validity"
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	c.Validity = t.String()
	// TODO - Validate candidate fields shouldn't be empty.
	c.Token = randStringBytes(33)
	m := add(c)
	mr := dgraph.SendMutation(m)
	if mr.Code != "ErrorOk" {
		sr.Write(w, mr.Message, "Mutation couldn't be applied by Dgraph.",
			http.StatusInternalServerError)
		return
	}

	// mutation applied successfully, lets send a mail to the candidate.
	uid, ok := mr.Uids["c"]
	if !ok {
		sr.Write(w, "Uid not returned for newly created candidate by Dgraph.",
			"", http.StatusInternalServerError)
		return
	}

	// Token sent in mail is uid + the random string.
	go mail.Send(c.Name, c.Email, uid+c.Token)
	sr.Message = "Candidate added successfully."
	sr.Success = true
	w.Write(server.MarshalResponse(sr))
}

func edit(c Candidate) string {
	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + c.Uid + `> <email> "` + c.Email + `" . `)
	m.Set(`<_uid_:` + c.Uid + `> <name> "` + c.Name + `" . `)
	m.Set(`<_uid_:` + c.Uid + `> <validity> "` + c.Validity + `" . `)

	// When the quiz for which candidate is invited is changed, we get both OldQuizId
	// and new QuizId.
	if c.QuizId != "" {
		m.Set(`<_uid_:` + c.QuizId + `> <quiz.candidate> <_uid_:` + c.Uid + `> .`)
		m.Set(`<_uid_:` + c.Uid + `> <candidate.quiz> <_uid_:` + c.QuizId + `> .`)
	}
	if c.OldQuizId != "" {
		m.Del(`<_uid_:` + c.OldQuizId + `> <quiz.candidate> <_uid_:` + c.Uid + `> .`)
		m.Del(`<_uid_:` + c.Uid + `> <candidate.quiz> <_uid_:` + c.OldQuizId + `> .`)
	}

	return m.String()
}

// TODO - Changing the quiz for a candidate doesn't work right now. Fix it.
func Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["id"]
	var c Candidate
	sr := server.Response{}
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't decode JSON", http.StatusBadRequest)
		return
	}

	var t time.Time
	if t, err = time.Parse(validityLayout, c.Validity); err != nil {
		sr.Message = "Couldn't parse the validity"
		sr.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.MarshalResponse(sr))
		return
	}

	c.Uid = cid
	c.Validity = t.String()
	// TODO - Validate candidate fields shouldn't be empty.
	m := edit(c)
	res := dgraph.SendMutation(m)
	if res.Code != "ErrorOk" {
		sr.Write(w, res.Message, "Mutation couldn't be applied by Dgraph.",
			http.StatusInternalServerError)
		return
	}
	go mail.Send(c.Name, c.Email, c.Uid+c.Token)
	sr.Success = true
	sr.Message = "Candidate info updated successfully."
	w.Write(server.MarshalResponse(sr))
}

func get(candidateId string) string {
	return `
    {
	quiz.candidate(_uid_:` + candidateId + `) {
		name
		email
		token
		validity
		complete
		candidate.quiz {
			_uid_
			duration
		}
	  }
    }`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["id"]
	q := get(cid)
	res := dgraph.Query(q)
	w.Write(res)
}

type quiz struct {
	Id        string           `json:"_uid_"`
	Duration  string           `json:"duration"`
	Questions []quizp.Question `json:"quiz.question"`
}

type qnIdsResp struct {
	Quizzes []quiz `json:"quiz"`
}

func quizQns(quizId string, qnsAsked []string) ([]quizp.Question, error) {
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
	res := dgraph.Query(q)
	var resp qnIdsResp
	json.Unmarshal(res, &resp)
	if len(resp.Quizzes) != 1 {
		return []quizp.Question{}, fmt.Errorf("Expected length of quizzes: %v. Got %v",
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

type resp struct {
	Cand []Candidate `json:"quiz.candidate"`
}

type Res struct {
	Token    string `json:"token"`
	Duration string `json:"duration"`
	Started  bool   `json:"quiz_started"`
	Name     string
}

func qnsAsked(qns []qids) []string {
	var uids []string
	for _, qn := range qns {
		uids = append(uids, qn.QuestionUid[0].Id)
	}
	return uids
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

	c, err := quizp.ReadMap(uid)
	// Check for duplicate session.
	if err == nil && !c.LastExchange().IsZero() {
		timeSinceLastExchange := time.Now().Sub(c.LastExchange())
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
	res := dgraph.Query(q)
	var resp resp
	json.Unmarshal(res, &resp)
	if len(resp.Cand) != 1 || len(resp.Cand[0].Quiz) != 1 {
		// No candidiate found with given uid
		sr.Write(w, "", "Invalid token.", http.StatusUnauthorized)
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

	if timeLeft(c.QuizStart(), dur) < 0 {
		sr.Write(w, "", "Your token is no longer valid.", http.StatusUnauthorized)
		return
	}

	if resp.Cand[0].Complete {
		sr.Write(w, "", "You have already completed the quiz.",
			http.StatusUnauthorized)
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
	// TODO - Shuffle the order of questions.
	// x.Shuffle(ids)

	if len(cand.Questions) > 0 {
		quizp.Update(uid, qns)
	} else {
		quizp.New(uid, qns, dur)
	}

	// Add the user id as a claim and return the token.
	claims := x.Claims{
		UserId: uid,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := jwtToken.SignedString([]byte(*auth.Secret))
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	// TODO - Incase candidate already has a active session return error after
	// implementing Ping.
	json.NewEncoder(w).Encode(Res{
		Token:    tokenString,
		Duration: timeLeft(c.QuizStart(), dur).String(),
		// Whether quiz was already started by the candidate.
		// If this is true the client can just call the questions API and
		// skip showing the instructions page.
		Started: len(cand.Questions) > 0,
		Name:    cand.Name,
	})
}
