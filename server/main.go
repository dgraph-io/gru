package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/dgraph-io/gru/server/interact"
	"google.golang.org/grpc"

	"gopkg.in/yaml.v2"
)

const (
	DEMO     = "demo"
	TEST     = "test"
	END      = "END"
	CORRECT  = 1
	WRONG    = 2
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
	logFile      *os.File
	testStart    time.Time
	// session id of currently active session.
	sid string
}

var (
	quizFile   = flag.String("quiz", "test.yml", "Input question file")
	port       = flag.String("port", ":8888", "Port on which server listens")
	candFile   = flag.String("cand", "candidates.txt", "Candidate inforamation")
	maxDemoQns = 3
	questions  []Question
	cmap       map[string]Candidate
	// List of question ids.
	qnList     []string
	demoQnList []string
	wrtLock    sync.Mutex
)

type server struct{}

func checkToken(c Candidate) error {
	if time.Now().UTC().After(c.validity) {
		return errors.New("Your token has expired.")
	}
	// Initially testStart is zero, but after candidate has taken the
	// test once, it shouldn't be zero.
	if !c.testStart.IsZero() && time.Now().UTC().After(c.testStart.Add(DURATION)) {
		return errors.New(fmt.Sprintf("%v since you started the test for the first time are already over.",
			DURATION))
	}
	return nil
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

func (c Candidate) loadCandInfo(token string) error {
	f, err := os.Open(fmt.Sprintf("logs/%s.log", token))
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
			score += float32(s)
			if len(qnsAsked) > 0 && splits[4] == qnsAsked[len(qnsAsked)-1] {
				continue
			}
			qnsAsked = append(qnsAsked, splits[4])
		case "ping":
			if len(qnsAsked) > 0 && splits[4] == qnsAsked[len(qnsAsked)-1] {
				continue
			}
			qnsAsked = append(qnsAsked, splits[4])
		}
	}
	c.score = score
	c.logFile = f
	c.questions = sliceDiff(questions, qnsAsked)
	// TODO(pawan) - Extract timeLeft from logs too. Maybe send initial time
	// from server when test starts
	cmap[token] = c
	return nil
}

func populateCandInfo(c Candidate, token string) {
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
		cmap[token] = c
		return
	}
	// If file exists but the test start is zero means the server might have
	// lost the data in memory, so we need to load it back from the file.
	// if c.testStart.IsZero() {
	// 	c.loadCandInfo(token)
	// }
}

// TODO - Don't allow multiple sessions simultaneously.
func authenticate(t *interact.Token) (*interact.Session, error) {
	var c Candidate
	var ok bool

	// This indicates there is no entry in the candidate file with this token.
	if c, ok = cmap[t.Id]; !ok {
		return nil, errors.New("Invalid token.")
	}

	populateCandInfo(c, t.Id)
	c = cmap[t.Id]
	if err := checkToken(c); err != nil {
		return nil, err
	}

	session := interact.Session{Id: RandStringBytes(36)}
	writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), session.Id))
	c.sid = session.Id
	cmap[t.Id] = c
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
				cmap[token] = c
				return formQuestion(q, c.score), nil
			}
		}
	}
	return &interact.Question{},
		fmt.Errorf("Didn't find qn with label: %s, for candidate: %s", qnType,
			token)
}

func getQuestion(req *interact.Req) (*interact.Question, error) {
	var c Candidate
	var ok bool
	if c, ok = cmap[req.Token]; !ok {
		return nil, errors.New("Invalid token.")
	}

	if c.demoQnsAsked < maxDemoQns {
		q, err := nextQuestion(c, req.Token, DEMO)
		if err != nil {
			return nil, err
		}
		return q, nil
	}
	// TOOD(pawan) - Check if time is up
	if len(c.questions) == 0 {
		q := &interact.Question{Id: END, Totscore: c.score}
		c.logFile.Close()
		return q, nil
	}
	// This means its the first test question.
	if len(c.questions) == len(questions)-maxDemoQns && !c.testStart.IsZero() {
		writeLog(c, fmt.Sprintf("%v test_start\n", UTCTime()))
		c.testStart = time.Now()
		return &interact.Question{Id: "DEMOEND"}
	}
	q, err := nextQuestion(c, req.Token, TEST)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func isValidSession(token string, sid string) error {
	var c Candidate
	var ok bool

	if c, ok = cmap[token]; !ok {
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

func isCorrectAnswer(resp *interact.Response) (int, int64) {
	for idx, que := range questions {
		if que.Id == resp.Qid {
			if reflect.DeepEqual(resp.Aid, que.Correct) {
				return idx, CORRECT
			} else {
				return idx, WRONG
			}
		}
	}
	return -1, -1
}

func UTCTime() string {
	return time.Now().UTC().Format("2006/01/02 15:04:05 MST")
}

func writeLog(c Candidate, s string) {
	wrtLock.Lock()
	c.logFile.WriteString(s)
	wrtLock.Unlock()
}

func sendAnswer(resp *interact.Response) (*interact.Status, error) {
	var c Candidate
	var ok bool
	if c, ok = cmap[resp.Token]; !ok {
		return &interact.Status{Status: 0}, errors.New("Invalid token.")
	}

	var status interact.Status
	var err error
	var idx int

	idx, status.Status = isCorrectAnswer(resp)
	if idx == -1 {
		log.Fatalf("Didn't find question with Id, %v.", resp.Qid)
	}
	if len(resp.Aid) > 0 && resp.Aid[0] != "skip" {
		if status.Status == 1 {
			c.score += questions[idx].Positive
		} else {
			c.score -= questions[idx].Negative
		}
	} else {
		if len(resp.Aid) > 1 {
			log.Println("Got extra optoins with SKIP")
		}
	}
	cmap[resp.Token] = c
	writeLog(c, fmt.Sprintf("%s response %s %s %.1f\n", UTCTime(),
		resp.Qid, strings.Join(resp.Aid, ","), c.score))

	return &status, err
}

func (s *server) SendAnswer(ctx context.Context,
	resp *interact.Response) (*interact.Status, error) {
	if err := isValidSession(resp.Token, resp.Sid); err != nil {
		return &interact.Status{}, err
	}
	return sendAnswer(resp)
}

func streamSend(wg *sync.WaitGroup, stream interact.GruQuiz_StreamChanServer,
	c Candidate, endTT chan int) {
	var stat interact.ServerStatus
	endTimeChan := time.NewTimer(DURATION).C
	tickChan := time.NewTicker(time.Second * 5).C

	for {
		stat.TimeLeft = time.Now().Sub(c.testStart).String()

		select {
		case <-endTimeChan:
			{
				endTT <- 1
				stat.Status = "END"
				fmt.Println("End test based on time")
				writeLog(c, fmt.Sprintf("End of test. Time out\n"))
				if err := stream.Send(&stat); err != nil {
					endTT <- 2
					log.Printf("Error while sending stream: %v\n",
						err)
				}
				wg.Done()
			}
		case <-tickChan:
			{
				stat.Status = " ONGOING"
				if err := stream.Send(&stat); err != nil {
					endTT <- 2
					log.Printf("Error while sending stream: %v\n",
						err)
				}
			}
		}
	}
}

func streamRecv(wg *sync.WaitGroup, stream interact.GruQuiz_StreamChanServer,
	c Candidate, endTT chan int) {
	for {
		select {
		case x := <-endTT:
			if x == 1 {
				log.Println("Received End test token")
			} else if x == 2 {
				log.Println("Possible Client crash")

			}
			wg.Done()
			return
		default:
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("Error while receiving stream: %v\n", err)
				}
				wg.Done()
				return
			}
			writeLog(c, fmt.Sprintf("%v ping %s\n",
				UTCTime(), msg.CurrQuestion))

		}
	}
	wg.Done()
}

// TODO(ashwin) - Add authentication
func (s *server) StreamChan(stream interact.GruQuiz_StreamChanServer) error {

	endTT := make(chan int)
	var wg sync.WaitGroup

	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error while receiving stream %v", err)
	}
	token := msg.Token
	c := cmap[token]

	wg.Add(1)
	go streamSend(&wg, stream, c, endTT)
	wg.Add(1)
	go streamRecv(&wg, stream, c, endTT)
	wg.Wait()
	return nil
}

func runGrpcServer(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("Error running quiz server %v", err)
		return
	}
	log.Printf("Server listening on address: %v", ln.Addr())

	s := grpc.NewServer()
	interact.RegisterGruQuizServer(s, &server{})
	if err = s.Serve(ln); err != nil {
		log.Fatalf("While serving gRpc requests %v", err)
	}
	return
}

func parseCandidateFile(file string) error {
	cmap = make(map[string]Candidate)
	format := "2006/01/02 (MST)"
	f, err := os.Open(*candFile)
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

		var c Candidate
		splits := strings.Split(line, " ")
		if len(splits) < 6 {
			log.Fatalf("Candidate info isn't sufficient for line %v",
				line)
		}

		c.name = strings.Join(splits[:2], " ")
		c.email = splits[2]
		c.validity, err = time.Parse(format,
			fmt.Sprintf("%s (%s)", splits[3], splits[4]))
		if err != nil {
			log.Fatal(err)
		}

		token := splits[5]
		cmap[token] = c
	}
	return nil
}

func checkIds(qns []Question) error {
	qidMap := make(map[string]bool)
	aidMap := make(map[string]bool)

	for _, q := range qns {
		if _, ok := qidMap[q.Id]; ok {
			return fmt.Errorf("Qn Id has been used before: %v", q.Id)
		}
		qidMap[q.Id] = true
		for _, ans := range q.Opt {
			if _, ok := aidMap[ans.Uid]; ok {
				return fmt.Errorf("Ans Id has been used before: %v",
					ans.Uid)
			}
			aidMap[ans.Uid] = true
		}
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
	err = checkIds(info)
	if err != nil {
		return []Question{}, err
	}
	return info, nil
}

func main() {
	flag.Parse()
	var err error
	if questions, err = extractQuizInfo(*quizFile); err != nil {
		log.Fatal(err)
	}
	parseCandidateFile(*candFile)
	// TODO(pawan) - Read testStart timings for candidates.
	runGrpcServer(*port)
}
