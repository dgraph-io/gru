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
	DEMO     = "demo"
	TEST     = "test"
	END      = "END"
	DURATION = 60 * time.Minute
)

type Candidate struct {
	name     string
	email    string
	validity time.Time
	score    float32
	// List of test questions that have not been asked to the candidate yet.
	questions []Question
	// count of demo qns asked.
	demoQnsAsked int
	demoTaken    bool
	logFile      *os.File
	demoStart    time.Time
	testStart    time.Time
	// session id of currently active session.
	sid          string
	lastExchange time.Time
}

var (
	tls        = flag.Bool("tls", true, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "fullchain.pem", "The TLS cert file")
	keyFile    = flag.String("key_file", "privkey.pem", "The TLS key file")
	quizFile   = flag.String("quiz", "test.yml", "Input question file")
	port       = flag.String("port", ":443", "Port on which server listens")
	candFile   = flag.String("cand", "candidates.txt", "Candidate inforamation")
	maxDemoQns = 8
	// List of question ids.
	questions []Question
	cmap      map[string]Candidate
	wrtLock   sync.Mutex
	mapLock   sync.Mutex
)

type server struct{}

func checkToken(c Candidate) error {
	if time.Now().UTC().After(c.validity) {
		return errors.New("Your token has expired.")
	}
	// Initially testStart is zero, but after candidate has taken the
	// test once, it shouldn't be zero.
	if !c.testStart.IsZero() && time.Now().UTC().After(c.testStart.Add(DURATION)) {
		// TODO - Show duration elapsed in minutes.
		return errors.New(fmt.Sprintf(
			"%v since you started the test for the first time are already over.",
			DURATION))
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
		case "test_start":
			c.testStart, err = time.Parse(format, fmt.Sprintf("%s %s %s",
				splits[0], splits[1], splits[2]))
			if err != nil {
				return err
			}
		case "response":
			s, err := strconv.ParseFloat(splits[6], 32)
			if err != nil {
				return err
			}
			// We only want to add score from actual quiz questions
			// and not demo qns.
			if len(qnsAsked) > maxDemoQns {
				score += float32(s)
			}
		case "question":
			qnsAsked = append(qnsAsked, splits[4])
		}
	}
	c.score = score
	c.logFile = f
	c.questions = sliceDiff(questions, qnsAsked)
	if len(questions)-len(c.questions) >= maxDemoQns {
		c.demoQnsAsked = maxDemoQns
		c.demoTaken = true
	} else {
		c.demoQnsAsked = len(questions) - len(c.questions)
	}
	return nil
}

func candInfo(token string) Candidate {
	// This indicates candidate info exists in memory and the client could
	// have crashed.
	c, _ := readMap(token)
	// We don't want to load up cand info for dummy test candidates.
	if strings.HasPrefix(token, "test-") {
		return c
	}
	if len(c.questions) > 0 || c.demoTaken {
		return c
	}

	// If file for the token doesn't exist means client is trying to connect
	// for the first time. So we create a file
	if _, err := os.Stat(fmt.Sprintf("logs/%s.log", token)); os.IsNotExist(err) {
		f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", token),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		c.logFile = f
		c.questions = make([]Question, len(questions))
		copy(c.questions, questions)
		updateMap(token, c)
		return c
	}
	// If we reach here it means logfile for candidate exists but his info
	// doesn't exist in memory, so we need to load it back from the file.
	// TODO - Don't allow multiple sessions simultaneously.
	var err error
	err = c.loadCandInfo(token)
	if err != nil {
		log.Fatalf("error while reading candidate info from log file,token: %s",
			token)
	}
	updateMap(token, c)
	return c
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func onlyDemoQuestions() []Question {
	var que []Question
	count := 0
	for _, x := range questions {
		if stringInSlice("demo", x.Tags) && count < maxDemoQns {
			que = append(que, x)
			count++
		}
	}
	return que
}

func demoCandInfo(token string) Candidate {
	var c Candidate
	c.name = token
	c.email = "no-mail@given"
	c.validity = time.Now().Add(time.Duration(100 * time.Hour))

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

func state(c Candidate) interact.QuizState {
	if len(c.questions) == len(questions) {
		return interact.Quiz_DEMO_NOT_TAKEN
	}
	if len(questions)-len(c.questions) < maxDemoQns {
		return interact.Quiz_DEMO_STARTED
	}
	if len(questions)-len(c.questions) == maxDemoQns {
		return interact.Quiz_TEST_NOT_TAKEN
	}
	if len(c.questions) == 0 {
		return interact.Quiz_TEST_FINISHED
	}
	return interact.Quiz_TEST_STARTED
}

func authenticate(t *interact.Token) (*interact.Session, error) {
	var c Candidate
	var ok bool
	var session interact.Session

	if strings.HasPrefix(t.Id, "test-") {
		c = demoCandInfo(t.Id)
		session = interact.Session{Id: RandStringBytes(36), State: interact.Quiz_DEMO_NOT_TAKEN,
			TimeLeft:     timeLeft(c.testStart).String(),
			TestDuration: DURATION.String()}
		return &session, nil
	}
	if c, ok = readMap(t.Id); !ok {
		return nil, errors.New("Invalid token.")
	}
	c = candInfo(t.Id)
	if err := checkToken(c); err != nil {
		return nil, err
	}

	timeSinceLastExchange := time.Now().Sub(c.lastExchange)
	if !c.lastExchange.IsZero() && timeSinceLastExchange < 10*time.Second {
		fmt.Println("Duplicate session for same auth token", t.Id, c.name)
		return nil, errors.New("Duplicate Session. You already have an open session. If not try after 10 seconds.")
	}

	session = interact.Session{Id: RandStringBytes(36), State: state(c),
		TimeLeft:     timeLeft(c.testStart).String(),
		TestDuration: DURATION.String()}
	writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), session.Id))
	c.sid = session.Id
	c.lastExchange = time.Now()
	updateMap(t.Id, c)
	return &session, nil
}

func (s *server) Authenticate(ctx context.Context,
	t *interact.Token) (*interact.Session, error) {
	return authenticate(t)
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
		Totscore: score}
}

func nextQuestion(c Candidate, token string, qnType string) (*interact.Question,
	error) {
	for idx, q := range c.questions {
		for _, t := range q.Tags {
			// For now qnType can just be "demo" or "test", later
			// it would have the difficulity level too.
			if qnType == t {
				c.questions = append(c.questions[:idx],
					c.questions[idx+1:]...)
				if qnType == DEMO {
					c.demoQnsAsked += 1
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
	var c Candidate
	var ok bool
	if c, ok = readMap(req.Token); !ok {
		return nil, errors.New("Invalid token.")
	}

	c.lastExchange = time.Now()
	updateMap(req.Token, c)

	if c.demoQnsAsked < maxDemoQns {
		if c.demoQnsAsked == 0 {
			c.demoStart = time.Now().UTC()
			writeLog(c, fmt.Sprintf("%v demo_start\n", UTCTime()))
		}
		q, err := nextQuestion(c, req.Token, DEMO)
		if err != nil {
			return nil, err
		}
		writeLog(c, fmt.Sprintf("%v question %v\n", UTCTime(), q.Id))
		return q, nil
	}

	if c.demoQnsAsked == maxDemoQns && c.testStart.IsZero() {
		if !c.demoTaken {
			c.demoTaken = true
			updateMap(req.Token, c)
			return &interact.Question{Id: "DEMOEND", Totscore: c.score},
				nil
		}
		c.score = 0
		updateMap(req.Token, c)
		// This means it is his first test question.
		writeLog(c, fmt.Sprintf("%v test_start\n", UTCTime()))
		c.testStart = time.Now().UTC()
	}
	q, err := nextQuestion(c, req.Token, TEST)
	// This means that test qns are over.
	if err != nil {
		q = &interact.Question{Id: END, Totscore: c.score}
		writeLog(c, fmt.Sprintf("%v End of test. Questions over\n", UTCTime()))
		return q, nil
	}
	writeLog(c, fmt.Sprintf("%v question %v\n", UTCTime(), q.Id))
	return q, nil
}

func isValidSession(token string, sid string) error {
	var c Candidate
	var ok bool

	if c, ok = readMap(token); !ok {
		return errors.New("Invalid token.")
	}

	if c.sid != "" && c.sid != sid {
		return errors.New("You already have another session active.")
	}
	return nil
}

func (s *server) GetQuestion(ctx context.Context,
	req *interact.Req) (*interact.Question, error) {
	if err := isValidSession(req.Token, req.Sid); err != nil {
		return &interact.Question{}, err
	}
	return getQuestion(req)
}

func isCorrectAnswer(resp *interact.Response) (int, float32) {
	for idx, que := range questions {
		if que.Id == resp.Qid {
			if resp.Aid[0] == "skip" {
				return idx, 0
			}
			// Multiple choice questions.
			if len(que.Correct) > 1 {
				var score float32
				for _, aid := range resp.Aid {
					correct := false
					for _, caid := range que.Correct {
						if caid == aid {
							correct = true
							break
						}
					}
					if correct {
						score += que.Positive
					} else {
						score -= que.Negative
					}
				}
				return idx, score
			}
			if resp.Aid[0] == que.Correct[0] {
				return idx, que.Positive
			} else {
				return idx, -que.Negative
			}
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

func sendAnswer(resp *interact.Response) (*interact.Status, error) {
	var c Candidate
	var ok bool
	var status interact.Status
	if c, ok = readMap(resp.Token); !ok {
		return &status, errors.New("Invalid token.")
	}

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
	writeLog(c, fmt.Sprintf("%s response %s %s %.1f\n", UTCTime(),
		resp.Qid, strings.Join(resp.Aid, ","), score))
	if idx == -1 {
		log.Printf("Didn't find qn: %v, token: %v, session: %v",
			resp.Qid, resp.Token, resp.Sid)
	}
	c.score += score
	updateMap(resp.Token, c)
	return &status, nil
}

func (s *server) SendAnswer(ctx context.Context,
	resp *interact.Response) (*interact.Status, error) {
	if err := isValidSession(resp.Token, resp.Sid); err != nil {
		return &interact.Status{}, err
	}
	return sendAnswer(resp)
}

func timeLeft(ts time.Time) time.Duration {
	return (DURATION - time.Now().UTC().Sub(ts))
}

func (s *server) Ping(ctx context.Context,
	stat *interact.ClientStatus) (*interact.ServerStatus, error) {
	var sstat interact.ServerStatus
	var c Candidate
	var ok bool

	if c, ok = readMap(stat.Token); !ok {
		return &sstat, fmt.Errorf("Invalid token: %v", stat.Token)
	}
	// In case the server crashed, we need to load up candidate info as
	// authenticate call won't be made.
	c = candInfo(stat.Token)

	writeLog(c, fmt.Sprintf("%v ping %s\n",
		UTCTime(), stat.CurrQuestion))

	c.lastExchange = time.Now()
	updateMap(stat.Token, c)

	if c.demoStart.IsZero() {
		log.Printf("Got ping before demo for Cand: %v", c.name)
		return &sstat, nil
	}
	// We want to indicate end of demo and test based on time.
	if c.testStart.IsZero() {
		demoTimeLeft := timeLeft(c.demoStart)
		if demoTimeLeft > 0 {
			sstat.TimeLeft = demoTimeLeft.String()
			return &sstat, nil
		}
		// So that now actual test questions are asked.
		c.demoQnsAsked = maxDemoQns
		c.demoTaken = true
		updateMap(stat.Token, c)
		sstat.Status = "DEMOEND"
		return &sstat, nil
	}
	quizTimeLeft := timeLeft(c.testStart)
	if quizTimeLeft > 0 {
		sstat.TimeLeft = quizTimeLeft.String()
		return &sstat, nil
	}
	sstat.Status = "END"
	writeLog(c, fmt.Sprintf("%v test_end\n", UTCTime()))
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

		// If the token already exists, skip that line
		token := splits[5]
		if _, ok := readMap(token); ok {
			continue
		}

		var c Candidate
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

func checkTest(qns []Question) error {
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
		if len(q.Correct) > 1 && q.Negative < q.Positive {
			return fmt.Errorf("Negative score less than positive for multi-choice qn: %v",
				q.Id)
		}
		for _, corr := range q.Correct {
			if _, ok := idsMap[corr]; !ok {
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
	err = checkTest(info)
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

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()
	var err error
	cmap = make(map[string]Candidate)
	if questions, err = extractQuizInfo(*quizFile); err != nil {
		log.Fatal(err)
	}
	go parseCandRepeat(*candFile)
	runGrpcServer(*port)
}
