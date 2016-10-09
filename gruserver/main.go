/*
 * Copyright 2016 DGraph Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	quizmeta "github.com/dgraph-io/gru/gruserver/quiz"
)

const (
	demo         = "demo"
	quiz         = "quiz"
	end          = "END"
	demoEnd      = "DEMOEND"
	demoDuration = 20 * time.Minute
	quizDuration = 60 * time.Minute
	// This is the number of demo questions asked to actual quiz candidates.
	beforeQuiz = 3
)

type Candidate struct {
	name     string
	email    string
	validity time.Time
	score    float32
	// List of quiz questions that have not been asked to the candidate yet.
	questions []Question
	// count of demo qns asked.
	demoQnsAsked int
	demoQnsToAsk int
	demoTaken    bool
	logFile      *os.File
	demoStart    time.Time
	quizStart    time.Time
	// session id of currently active session.
	sid          string
	lastExchange time.Time
}

var (
	tls      = flag.Bool("tls", true, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "fullchain.pem", "The TLS cert file")
	keyFile  = flag.String("key_file", "privkey.pem", "The TLS key file")
	quizFile = flag.String("quiz", "quiz.yml", "Input question file")
	port     = flag.String("port", ":443", "Port on which server listens")
	candFile = flag.String("cand", "candidates.txt", "Candidate inforamation")
	// This is the number of demo questions asked to dummy candidates.
	maxDemoQns = flag.Int("max_demo_qns", 25, "Maximum number of demo questions for dummy candidates.")
	// List of question ids.
	questions []Question
	cmap      map[string]Candidate
	wrtLock   sync.Mutex
	mapLock   sync.Mutex
	throttle  = make(chan time.Time, 3)
	rate      = time.Second
)

type server struct{}

func checkToken(c Candidate) error {
	if time.Now().UTC().After(c.validity) {
		return errors.New("Your token has expired.")
	}
	// Initially quizStart is zero, but after candidate has taken the
	// quiz once, it shouldn't be zero.
	if !c.quizStart.IsZero() && time.Now().UTC().After(c.quizStart.Add(quizDuration)) {
		// TODO - Show duration elapsed in minutes.
		return fmt.Errorf(
			"%v since you started the quiz for the first time are already over",
			quizDuration)
	}
	return nil
}

func updateMap(token string, c Candidate) {
	mapLock.Lock()
	defer mapLock.Unlock()
	cmap[token] = c
}

func readMap(token string) (Candidate, bool) {
	mapLock.Lock()
	defer mapLock.Unlock()
	c, ok := cmap[token]
	return c, ok
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func sliceDiff(qnList []Question, qnsAsked []string) []Question {
	qns := []Question{}
	qmap := make(map[string]bool)
	for _, q := range qnsAsked {
		qmap[q] = true
	}
	for _, q := range qnList {
		if present := qmap[q.Id]; !present {
			qns = append(qns, q)
		}
	}
	return qns
}

func candInfo(token string, c Candidate) (Candidate, error) {
	// We don't want to load up cand info for dummy quiz candidates.
	if strings.HasPrefix(token, "test-") {
		return c, nil
	}
	if len(c.questions) > 0 || c.demoTaken {
		return c, nil
	}

	// If file for the token doesn't exist means client is trying to connect
	// for the first time. So we create a file
	if _, err := os.Stat(fmt.Sprintf("logs/%s.log", token)); os.IsNotExist(err) {
		f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", token),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return c, err
		}
		c.logFile = f
		c.questions = make([]Question, len(questions))
		copy(c.questions, questions)
		updateMap(token, c)
		return c, nil
	}

	// var err error
	// // If we reach here it means logfile for candidate exists but his info
	// // doesn't exist in memory, so we need to load it back from the file.
	// if err = c.loadCandInfo(token); err != nil {
	// 	return c, err
	// }
	updateMap(token, c)
	return c, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func shuffleOptions(opts []*quizmeta.Answer) {
	for i := range opts {
		j := rand.Intn(i + 1)
		opts[i], opts[j] = opts[j], opts[i]
	}
}

func shuffleQuestions(qns []Question) {
	for i := range qns {
		j := rand.Intn(i + 1)
		qns[i], qns[j] = qns[j], qns[i]
	}
}

func onlyDemoQuestions() []Question {
	var qns []Question
	count := 0
	for _, x := range questions {
		if stringInSlice("demo", x.Tags) && count < *maxDemoQns {
			qns = append(qns, x)
			count++
		}
	}
	shuffleQuestions(qns)
	return qns
}

func demoCandInfo(token string) Candidate {
	var c Candidate
	c.name = token
	c.email = "no-mail@given"
	c.validity = time.Now().Add(time.Duration(100 * time.Hour))
	c.demoQnsToAsk = *maxDemoQns

	if _, err := os.Stat(fmt.Sprintf("logs/%s.log", token)); os.IsNotExist(err) {
		f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", token),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		c.logFile = f
		c.questions = onlyDemoQuestions()
		updateMap(token, c)
		return c
	}
	updateMap(token, c)
	return c
}

type QuizState int32

const (
	QUIZ_DEMO_NOT_TAKEN QuizState = 0
	QUIZ_DEMO_STARTED   QuizState = 1
	QUIZ_TEST_NOT_TAKEN QuizState = 2
	QUIZ_TEST_STARTED   QuizState = 3
	QUIZ_TEST_FINISHED  QuizState = 4
)

func state(c Candidate) QuizState {
	if len(c.questions) == len(questions) {
		return QUIZ_DEMO_NOT_TAKEN
	}
	if len(questions)-len(c.questions) < c.demoQnsToAsk {
		return QUIZ_DEMO_STARTED
	}
	if len(questions)-len(c.questions) == c.demoQnsToAsk {
		return QUIZ_TEST_NOT_TAKEN
	}
	if len(c.questions) == 0 {
		return QUIZ_TEST_FINISHED
	}
	return QUIZ_TEST_STARTED
}

type Session struct {
	Id           string    `json:"id"`
	State        QuizState `json:"state"`
	TimeLeft     string    `json:"timeleft"`
	TestDuration string    `json:"testduration"`
	DemoDuration string    `json:"demoduration"`
}

func authenticate(token string) (s Session, err error) {
	var c Candidate
	var ok bool
	if strings.HasPrefix(token, "test-") {
		c = demoCandInfo(token)
		s = Session{
			Id:           RandStringBytes(36),
			State:        QUIZ_DEMO_NOT_TAKEN,
			TestDuration: quizDuration.String(),
			DemoDuration: demoDuration.String(),
		}
		if !c.demoStart.IsZero() {
			s.TimeLeft = timeLeft(demoDuration, c.demoStart).String()
		} else {
			s.TimeLeft = demoDuration.String()
		}
		writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), s.Id))
		return
	}
	if c, ok = readMap(token); !ok {
		err = errors.New("Invalid token")
		return
	}

	if c, err = candInfo(token, c); err != nil {
		return
	}
	if err = checkToken(c); err != nil {
		return
	}

	timeSinceLastExchange := time.Now().Sub(c.lastExchange)
	if !c.lastExchange.IsZero() && timeSinceLastExchange < 10*time.Second {
		fmt.Println("Duplicate session for same auth token", token, c.name)
		err = errors.New("Duplicate Session. You already have an open session. If not try after 10 seconds.")
		return
	}

	s = Session{
		Id:           RandStringBytes(36),
		State:        state(c),
		TestDuration: quizDuration.String(),
		DemoDuration: demoDuration.String(),
	}

	if state(c) == QUIZ_DEMO_NOT_TAKEN || state(c) == QUIZ_DEMO_STARTED {
		s.TimeLeft = timeLeft(demoDuration, c.demoStart).String()
	} else {
		s.TimeLeft = timeLeft(quizDuration, c.quizStart).String()
	}
	writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), s.Id))
	c.sid = s.Id
	c.lastExchange = time.Now()
	updateMap(token, c)
	return
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	var s Session
	var err error

	select {
	case <-throttle:
		s, err = authenticate(r.Header.Get("Authorization"))
	case <-time.After(time.Second * 1):
		err = errors.New("Please try again later. Too much load on server.")
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s)
}

type Option struct {
	Uid string
	Str string
}

type Question struct {
	Id       string
	Str      string
	Correct  []string
	Opt      []Option
	Positive float32
	Negative float32
	Tags     []string
}

func formQuestion(q Question, score float32) *quizmeta.Question {
	var opts []*quizmeta.Answer
	for _, o := range q.Opt {
		a := &quizmeta.Answer{Id: o.Uid, Str: o.Str}
		opts = append(opts, a)
	}
	shuffleOptions(opts)
	var isM bool
	if len(q.Correct) > 1 {
		isM = true
	}
	return &quizmeta.Question{Id: q.Id, Str: q.Str, Options: opts,
		IsMultiple: isM, Positive: q.Positive, Negative: q.Negative,
		Score: score}
}

func nextQuestion(c Candidate, token string, qnType string) (*quizmeta.Question,
	error) {
	for idx, q := range c.questions {
		for _, t := range q.Tags {
			// For now qnType can just be "demo" or "quiz", later
			// it would have the difficulity level too.
			if qnType == t {
				c.questions = append(c.questions[:idx],
					c.questions[idx+1:]...)
				if qnType == demo {
					c.demoQnsAsked++
				}
				updateMap(token, c)
				return formQuestion(q, c.score), nil
			}
		}
	}
	return &quizmeta.Question{},
		fmt.Errorf("Didn't find qn with label: %s, for candidate: %s",
			qnType, token)
}

func getQuestion(token string) (*quizmeta.Question, error) {
	var q *quizmeta.Question

	c, _ := readMap(token)
	c.lastExchange = time.Now()
	updateMap(token, c)

	if c.demoQnsAsked < c.demoQnsToAsk {
		if c.demoQnsAsked == 0 {
			c.demoStart = time.Now().UTC()
			writeLog(c, fmt.Sprintf("%v demo_start\n", UTCTime()))
		}
		q, err := nextQuestion(c, token, demo)
		if err != nil {
			return q, err
		}
		writeLog(c, fmt.Sprintf("%v question %v\n", UTCTime(), q.Id))
		return q, nil
	}

	if c.demoQnsAsked == c.demoQnsToAsk && c.quizStart.IsZero() {
		if !c.demoTaken {
			c.demoTaken = true
			updateMap(token, c)
			return &quizmeta.Question{Id: demoEnd,
				Score: c.score}, nil
		}
		c.score = 0
		updateMap(token, c)
		// This means it is his first quiz question.
		writeLog(c, fmt.Sprintf("%v quiz_start\n", UTCTime()))
		c.quizStart = time.Now().UTC()
	}
	q, err := nextQuestion(c, token, quiz)
	// This means that quiz qns are over.
	if err != nil {
		q = &quizmeta.Question{Id: end, Score: c.score}
		writeLog(c, fmt.Sprintf("%v End of quiz. Questions over\n",
			UTCTime()))
		return q, nil
	}
	writeLog(c, fmt.Sprintf("%v question %v\n", UTCTime(), q.Id))
	return q, nil
}

func isValidSession(token string, sid string) (Candidate, error) {
	var c Candidate
	var ok bool

	if c, ok = readMap(token); !ok {
		return Candidate{}, fmt.Errorf("Invalid token.")
	}

	if c.sid != "" && c.sid != sid {
		return Candidate{}, fmt.Errorf("You already have another session active.")
	}
	return c, nil
}

func GetQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	token := r.Header.Get("Authorization")

	_, err := isValidSession(token, r.FormValue("sid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	q, err := getQuestion(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func isCorrectAnswer(qid string, aids []string) (int, float32) {
	for idx, qn := range questions {
		if qn.Id == qid {
			if aids[0] == "skip" {
				return idx, 0
			}
			// For multiple choice qnstions, we have partial scoring.
			if len(qn.Correct) == 1 {
				if aids[0] == qn.Correct[0] {
					return idx, qn.Positive
				}
				return idx, -qn.Negative
			}
			var score float32
			for _, aid := range aids {
				correct := false
				for _, caid := range qn.Correct {
					if caid == aid {
						correct = true
						break
					}
				}
				if correct {
					score += qn.Positive
				} else {
					score -= qn.Negative
				}
			}
			return idx, score
		}
	}
	return -1, 0
}

func UTCTime() string {
	return time.Now().UTC().Format("2006/01/02 15:04:05 MST")
}

func writeLog(c Candidate, s string) {
	wrtLock.Lock()
	_, err := c.logFile.WriteString(s)
	if err != nil {
		log.Printf("Error: %v while writing logs to file for Cand: %v",
			err, c.name)
	}
	wrtLock.Unlock()
}

func status(token string, sid string, qid string, aids []string) (*quizmeta.AnswerStatus, error) {
	var status quizmeta.AnswerStatus
	c, _ := readMap(token)
	c.lastExchange = time.Now()
	updateMap(token, c)
	if len(aids) == 0 {
		log.Printf("Got empty response for qn:%v, token: %v, session: %v",
			qid, token, sid)
		return &status, nil
	}

	if aids[0] == "skip" && len(aids) > 1 {
		log.Printf("Got extra options with SKIP for qn:%v, token: %v, session: %v",
			qid, token, sid)
	}
	idx, score := isCorrectAnswer(qid, aids)
	writeLog(c, fmt.Sprintf("%s response %s %s %.1f\n", UTCTime(), qid,
		strings.Join(aids, ","), score))
	if idx == -1 {
		log.Printf("Didn't find qn: %v, token: %v, session: %v",
			qid, token, sid)
	}
	c.score += score
	updateMap(token, c)
	return &status, nil
}

func Status(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	token := r.Header.Get("Authorization")
	sid := r.FormValue("sid")

	_, err := isValidSession(token, sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	q, err := status(token, sid, r.FormValue("qid"), r.Form["aid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func timeLeft(dur time.Duration, ts time.Time) time.Duration {
	return (dur - time.Now().UTC().Sub(ts))
}

func (s *server) Ping(ctx context.Context,
	stat *quizmeta.ClientStatus) (*quizmeta.ServerStatus, error) {
	var sstat quizmeta.ServerStatus
	var c Candidate
	var ok bool

	if c, ok = readMap(stat.Token); !ok {
		return &sstat, fmt.Errorf("Invalid token: %v", stat.Token)
	}

	var err error
	// In case the server crashed, we need to load up candidate info as
	// authenticate call won't be made.
	if c, err = candInfo(stat.Token, c); err != nil {
		return &sstat, err
	}

	writeLog(c, fmt.Sprintf("%v ping %s\n", UTCTime(), stat.CurQuestion))
	c.lastExchange = time.Now()
	updateMap(stat.Token, c)
	if c.demoStart.IsZero() {
		log.Printf("Got ping before demo for Cand: %v", c.name)
		return &sstat, nil
	}

	// We want to indicate end of demo and quiz based on time. If quiz did
	// not start yet we send time left for demo.
	if c.quizStart.IsZero() {
		demoTimeLeft := timeLeft(demoDuration, c.demoStart)
		if demoTimeLeft > 0 {
			sstat.TimeLeft = demoTimeLeft.String()
			return &sstat, nil
		}
		// So that now actual quiz questions are asked.
		c.demoQnsAsked = beforeQuiz
		c.demoTaken = true
		updateMap(stat.Token, c)
		sstat.Status = demoEnd
		return &sstat, nil
	}

	quizTimeLeft := timeLeft(quizDuration, c.quizStart)
	if quizTimeLeft > 0 {
		sstat.TimeLeft = quizTimeLeft.String()
		return &sstat, nil
	}

	sstat.Status = end
	writeLog(c, fmt.Sprintf("%v quiz_end\n", UTCTime()))
	return &sstat, nil
}

func runHTTPServer(address string) {
	http.HandleFunc("/authenticate", Authenticate)
	http.HandleFunc("/nextquestion", GetQuestion)
	http.HandleFunc("/status", Status)
	log.Fatal(http.ListenAndServe(address, nil))
}

// This method is used to parse the candidate file which contains information
// about the candidates allowed to take the quiz.
func parseCandidateFile(file string) error {
	format := "2006/01/02 (MST)"
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == ';' {
			continue
		}

		splits := strings.Split(line, " ")
		if len(splits) < 6 {
			log.Fatalf("Candidate info isn't sufficient for line %v",
				line)
		}

		// If the token already exists, skip that line.
		token := splits[5]
		if _, ok := readMap(token); ok {
			continue
		}

		var c Candidate
		c.demoQnsToAsk = beforeQuiz
		c.name = strings.Join(splits[:2], " ")
		c.email = splits[2]
		c.validity, err = time.Parse(format,
			fmt.Sprintf("%s (%s)", splits[3], splits[4]))
		if err != nil {
			log.Fatal(err)
		}

		updateMap(token, c)
	}
	return nil
}

func partOfOptions(opts []Option, s string) bool {
	for _, opt := range opts {
		if s == opt.Uid {
			return true
		}
	}
	return false
}

// This method performs sanity checks on the data in the quiz file.
func checkQuiz(qns []Question) error {
	// None of the ids should be repeated, so we check that using a map.
	idsMap := make(map[string]bool)
	demoQnCount := 0

	for _, q := range qns {
		if _, ok := idsMap[q.Id]; ok {
			return fmt.Errorf("Id has been used before: %v", q.Id)
		}

		if stringInSlice("demo", q.Tags) {
			demoQnCount++
		}

		idsMap[q.Id] = true
		for _, ans := range q.Opt {
			if _, ok := idsMap[ans.Uid]; ok {
				return fmt.Errorf("Id has been used before: %v",
					ans.Uid)
			}
			idsMap[ans.Uid] = true
		}

		for _, tag := range q.Tags {
			if tag[0] < 'a' || tag[0] > 'z' {
				return fmt.Errorf(
					"Tag: %v for qn: %v should start with a lowercase character",
					tag, q.Id)
			}
		}

		if len(q.Correct) == 0 {
			return fmt.Errorf("Correct list is empty")
		}

		if q.Negative < 0 || q.Positive < 0 {
			return fmt.Errorf("Score for qn: %v is less than zero.",
				q.Id)
		}
		// As we do partial scoring for multiple questions, the negative
		// score shouldn't be less than positive score.
		if len(q.Correct) > 1 && q.Negative < q.Positive {
			return fmt.Errorf("Negative score less than positive for multi-choice qn: %v",
				q.Id)
		}

		for _, corr := range q.Correct {
			if ok := partOfOptions(q.Opt, corr); !ok {
				return fmt.Errorf("Correct not part of options: %v ",
					corr)
			}
		}
	}
	if demoQnCount < *maxDemoQns {
		return fmt.Errorf("Need more demo questions in quiz file")
	}
	return nil
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

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()
	cmap = make(map[string]Candidate)
	go rateLimit()
	runHTTPServer(*port)
}
