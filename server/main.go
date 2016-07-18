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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/dgraph-io/gru/server/interact"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"gopkg.in/yaml.v2"
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
	maxDemoQns = 25
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

// Parses a candidate log file and loads information about him in memory.
func (c *Candidate) loadCandInfo(token string) error {
	f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", token),
		os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	var score float32
	qnsAsked := []string{}
	format := "2006/01/02 15:04:05 MST"
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		splits := strings.Split(line, " ")
		if len(splits) < 4 {
			log.Printf("Log for token: %v, line: %v has less than 4 words",
				token, line)
			continue
		}
		if splits[3] == "session_token" {
			continue
		}

		switch splits[3] {
		case "demo_start":
			c.demoStart, err = time.Parse(format, fmt.Sprintf("%s %s %s",
				splits[0], splits[1], splits[2]))
			if err != nil {
				return err
			}
		case "quiz_start":
			c.quizStart, err = time.Parse(format, fmt.Sprintf("%s %s %s",
				splits[0], splits[1], splits[2]))
			if err != nil {
				return err
			}
		case "response":
			if len(splits) < 7 {
				log.Printf(
					"Response log for token: %v, line: %v has less than 7 words",
					token, line)
				continue
			}
			s, err := strconv.ParseFloat(splits[6], 32)
			if err != nil {
				return err
			}
			// We only want to add score from actual quiz questions
			// and not demo qns.
			if len(qnsAsked) > beforeQuiz {
				score += float32(s)
			}
		case "question":
			if len(splits) < 5 {
				log.Printf(
					"Question log for token: %v, line: %v has less than 5 words",
					token, line)
				continue
			}
			qnsAsked = append(qnsAsked, splits[4])
		}
	}
	c.score = score
	c.logFile = f
	c.questions = sliceDiff(questions, qnsAsked)
	if len(qnsAsked) >= beforeQuiz {
		c.demoQnsAsked = beforeQuiz
		c.demoTaken = true
	} else {
		c.demoQnsAsked = len(qnsAsked)
	}
	return nil
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

	var err error
	// If we reach here it means logfile for candidate exists but his info
	// doesn't exist in memory, so we need to load it back from the file.
	if err = c.loadCandInfo(token); err != nil {
		return c, err
	}
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

func shuffle(qns []Question) {
	for i := range qns {
		j := rand.Intn(i + 1)
		qns[i], qns[j] = qns[j], qns[i]
	}
}

func onlyDemoQuestions() []Question {
	var qns []Question
	count := 0
	for _, x := range questions {
		if stringInSlice("demo", x.Tags) && count < maxDemoQns {
			qns = append(qns, x)
			count++
		}
	}
	shuffle(qns)
	return qns
}

func demoCandInfo(token string) Candidate {
	var c Candidate
	c.name = token
	c.email = "no-mail@given"
	c.validity = time.Now().Add(time.Duration(100 * time.Hour))
	c.demoQnsToAsk = maxDemoQns

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

func state(c Candidate) interact.QUIZState {
	if len(c.questions) == len(questions) {
		return interact.QUIZ_DEMO_NOT_TAKEN
	}
	if len(questions)-len(c.questions) < c.demoQnsToAsk {
		return interact.QUIZ_DEMO_STARTED
	}
	if len(questions)-len(c.questions) == c.demoQnsToAsk {
		return interact.QUIZ_TEST_NOT_TAKEN
	}
	if len(c.questions) == 0 {
		return interact.QUIZ_TEST_FINISHED
	}
	return interact.QUIZ_TEST_STARTED
}

func authenticate(ctx context.Context,
	t *interact.Token) (*interact.Session, error) {
	if d, ok := ctx.Deadline(); ok && d.Before(time.Now()) {
		return &interact.Session{}, errors.New("Context deadline has passed.")
	}

	var c Candidate
	var ok bool
	var session interact.Session

	if strings.HasPrefix(t.Id, "test-") {
		c = demoCandInfo(t.Id)
		session = interact.Session{
			Id:           RandStringBytes(36),
			State:        interact.QUIZ_DEMO_NOT_TAKEN,
			TimeLeft:     timeLeft(demoDuration, c.demoStart).String(),
			TestDuration: quizDuration.String(),
			DemoDuration: demoDuration.String(),
		}
		return &session, nil
	}
	if c, ok = readMap(t.Id); !ok {
		return nil, errors.New("Invalid token.")
	}

	var err error
	if c, err = candInfo(t.Id, c); err != nil {
		return &session, err
	}
	if err = checkToken(c); err != nil {
		return &session, err
	}

	timeSinceLastExchange := time.Now().Sub(c.lastExchange)
	if !c.lastExchange.IsZero() && timeSinceLastExchange < 10*time.Second {
		fmt.Println("Duplicate session for same auth token", t.Id, c.name)
		return nil, errors.New("Duplicate Session. You already have an open session. If not try after 10 seconds.")
	}

	session = interact.Session{
		Id:           RandStringBytes(36),
		State:        state(c),
		TestDuration: quizDuration.String(),
		DemoDuration: demoDuration.String(),
	}

	if state(c) == interact.QUIZ_DEMO_NOT_TAKEN || state(c) == interact.QUIZ_DEMO_STARTED {
		session.TimeLeft = timeLeft(demoDuration, c.demoStart).String()
	} else {
		session.TimeLeft = timeLeft(quizDuration, c.quizStart).String()
	}
	writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), session.Id))
	c.sid = session.Id
	c.lastExchange = time.Now()
	updateMap(t.Id, c)
	return &session, nil
}

func (s *server) Authenticate(ctx context.Context,
	t *interact.Token) (*interact.Session, error) {

	select {
	case <-throttle:
		return authenticate(ctx, t)
	case <-time.After(time.Second * 1):
		return nil, errors.New("Please try again later. Too much load on server.")
	}
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

func formQuestion(q Question, score float32) *interact.Question {
	var opts []*interact.Answer
	for _, o := range q.Opt {
		a := &interact.Answer{Id: o.Uid, Str: o.Str}
		opts = append(opts, a)
	}

	var isM bool
	if len(q.Correct) > 1 {
		isM = true
	}
	return &interact.Question{Id: q.Id, Str: q.Str, Options: opts,
		IsMultiple: isM, Positive: q.Positive, Negative: q.Negative,
		Score: score}
}

func nextQuestion(c Candidate, token string, qnType string) (*interact.Question,
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
	return &interact.Question{},
		fmt.Errorf("Didn't find qn with label: %s, for candidate: %s",
			qnType, token)
}

func getQuestion(req *interact.Req) (*interact.Question, error) {
	c, _ := readMap(req.Token)
	c.lastExchange = time.Now()
	updateMap(req.Token, c)

	if c.demoQnsAsked < c.demoQnsToAsk {
		if c.demoQnsAsked == 0 {
			c.demoStart = time.Now().UTC()
			writeLog(c, fmt.Sprintf("%v demo_start\n", UTCTime()))
		}
		q, err := nextQuestion(c, req.Token, demo)
		if err != nil {
			return nil, err
		}
		writeLog(c, fmt.Sprintf("%v question %v\n", UTCTime(), q.Id))
		return q, nil
	}

	if c.demoQnsAsked == c.demoQnsToAsk && c.quizStart.IsZero() {
		if !c.demoTaken {
			c.demoTaken = true
			updateMap(req.Token, c)
			return &interact.Question{Id: demoEnd,
				Score: c.score}, nil
		}
		c.score = 0
		updateMap(req.Token, c)
		// This means it is his first quiz question.
		writeLog(c, fmt.Sprintf("%v quiz_start\n", UTCTime()))
		c.quizStart = time.Now().UTC()
	}
	q, err := nextQuestion(c, req.Token, quiz)
	// This means that quiz qns are over.
	if err != nil {
		q = &interact.Question{Id: end, Score: c.score}
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
		return Candidate{}, errors.New("Invalid token.")
	}

	if c.sid != "" && c.sid != sid {
		return Candidate{}, errors.New("You already have another session active.")
	}
	return c, nil
}

func (s *server) GetQuestion(ctx context.Context,
	req *interact.Req) (*interact.Question, error) {
	if d, ok := ctx.Deadline(); ok && d.Before(time.Now()) {
		return &interact.Question{}, errors.New("Context deadline has passed.")
	}
	_, err := isValidSession(req.Token, req.Sid)
	if err != nil {
		return &interact.Question{}, err
	}

	return getQuestion(req)
}

func isCorrectAnswer(resp *interact.Response) (int, float32) {
	for idx, qn := range questions {
		if qn.Id == resp.Qid {
			if resp.Aid[0] == "skip" {
				return idx, 0
			}
			// For multiple choice qnstions, we have partial scoring.
			if len(qn.Correct) == 1 {
				if resp.Aid[0] == qn.Correct[0] {
					return idx, qn.Positive
				}
				return idx, -qn.Negative
			}
			var score float32
			for _, aid := range resp.Aid {
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

func status(resp *interact.Response) (*interact.AnswerStatus, error) {
	var status interact.AnswerStatus

	c, _ := readMap(resp.Token)
	c.lastExchange = time.Now()
	updateMap(resp.Token, c)
	if len(resp.Aid) == 0 {
		log.Printf("Got empty response for qn:%v, token: %v, session: %v",
			resp.Qid, resp.Token, resp.Sid)
		return &status, nil
	}

	if resp.Aid[0] == "skip" && len(resp.Aid) > 1 {
		log.Printf("Got extra options with SKIP for qn:%v, token: %v, session: %v",
			resp.Qid, resp.Token, resp.Sid)
	}
	idx, score := isCorrectAnswer(resp)
	writeLog(c, fmt.Sprintf("%s response %s %s %.1f\n", UTCTime(), resp.Qid,
		strings.Join(resp.Aid, ","), score))
	if idx == -1 {
		log.Printf("Didn't find qn: %v, token: %v, session: %v",
			resp.Qid, resp.Token, resp.Sid)
	}
	c.score += score
	updateMap(resp.Token, c)
	return &status, nil
}

func (s *server) Status(ctx context.Context,
	resp *interact.Response) (*interact.AnswerStatus, error) {
	if d, ok := ctx.Deadline(); ok && d.Before(time.Now()) {
		return &interact.AnswerStatus{}, errors.New("Context deadline has passed.")
	}
	_, err := isValidSession(resp.Token, resp.Sid)
	if err != nil {
		return &interact.AnswerStatus{}, err
	}
	return status(resp)
}

func timeLeft(dur time.Duration, ts time.Time) time.Duration {
	return (dur - time.Now().UTC().Sub(ts))
}

func (s *server) Ping(ctx context.Context,
	stat *interact.ClientStatus) (*interact.ServerStatus, error) {
	if d, ok := ctx.Deadline(); ok && d.Before(time.Now()) {
		return &interact.ServerStatus{}, errors.New("Context deadline has passed.")
	}
	var sstat interact.ServerStatus
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

func runGrpcServer(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("Error running quiz server %v", err)
		return
	}
	log.Printf("Server listening on address: %v", ln.Addr())

	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}
	s := grpc.NewServer(opts...)
	interact.RegisterGruQuizServer(s, &server{})
	if err = s.Serve(ln); err != nil {
		log.Fatalf("While serving gRpc requests %v", err)
	}
	return
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
	if demoQnCount < maxDemoQns {
		return fmt.Errorf("Need more demo questions in quiz file")
	}
	return nil
}

// Reads the quiz file and converts the questions into the internal question format.
func extractQuizInfo(file string) ([]Question, error) {
	var info []Question
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return []Question{}, err
	}
	err = yaml.Unmarshal(b, &info)
	if err != nil {
		return []Question{}, err
	}
	err = checkQuiz(info)
	if err != nil {
		return []Question{}, err
	}
	return info, nil
}

func parseCandRepeat(file string) {
	parseCandidateFile(file)
	tickChan := time.NewTicker(time.Minute).C
	for {
		select {
		case <-tickChan:
			parseCandidateFile(file)
		}
	}
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
	var err error
	cmap = make(map[string]Candidate)
	if questions, err = extractQuizInfo(*quizFile); err != nil {
		log.Fatal(err)
	}
	go parseCandRepeat(*candFile)
	go rateLimit()
	runGrpcServer(*port)
}
