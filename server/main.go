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

	"google.golang.org/grpc"

	"github.com/dgraph-io/dgraph/x"
	"github.com/dgraph-io/gru/server/interact"
	"golang.org/x/net/context"
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

var (
	quizFile = flag.String("quiz", "test.yml", "Input question file")
	port     = flag.String("port", ":8888", "Port on which server listens")
	candFile = flag.String("cand", "candidates.txt", "Candidate inforamation")
	quizInfo map[string][]Question
	cmap     map[string]Candidate
	glog     = x.Log("Gru Server")
	// List of question ids.
	qnList     []string
	demoQnList []string
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

func sliceDiff(qnList, qnsAsked []string) []string {
	qns := []string{}
	qmap := make(map[string]bool)
	for _, q := range qnsAsked {
		qmap[q] = true
	}
	for _, q := range qnList {
		if present := qmap[q]; !present {
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
	if len(qnsAsked) > 0 {
		c.demoQnList = demoQnList[:]
	}
	c.qnList = sliceDiff(qnList, qnsAsked)
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
		c.qnList = qnList[:]
		c.demoQnList = demoQnList[:]
		cmap[token] = c
		return
	}
	// If file exists but the test start is zero means the server might have
	// lost the data in memory, so we need to load it back from the file.
	if c.testStart.IsZero() {
		c.loadCandInfo(token)
	}

}

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
	c.logFile.WriteString(fmt.Sprintf("%v session_token %s\n", UTCTime(),
		session.Id))
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
	Tag      string
}

func qnFromList(qid string, list []Question) Question {
	for _, q := range list {
		if q.Id == qid {
			return q
		}
	}
	return Question{}
}

func nextQuestion(c Candidate, list []string, testType string) (*interact.Question, []string) {
	idx := rand.Intn(len(list))
	qid := list[idx]
	q := qnFromList(qid, quizInfo[testType])

	var opts []*interact.Answer
	for _, o := range q.Opt {
		a := &interact.Answer{Id: o.Uid, Str: o.Str}
		opts = append(opts, a)
	}

	var isM bool
	if len(q.Correct) > 1 {
		isM = true
	}
	que := &interact.Question{Id: q.Id, Str: q.Str, Options: opts,
		IsMultiple: isM, Positive: q.Positive, Negative: q.Negative,
		Totscore: c.score,
	}
	list = append(list[:idx], list[idx+1:]...)
	return que, list
}

func getQuestion(req *interact.Req) (*interact.Question, error) {
	var c Candidate
	var ok bool
	if c, ok = cmap[req.Token]; !ok {
		return nil, errors.New("Invalid token.")
	}

	testType := req.TestType
	if testType == DEMO {
		if len(c.demoQnList) == 0 {
			// If demo qns are over, indicate end of demo to client.
			q := &interact.Question{Id: END, Totscore: 0}
			c.score = 0
			cmap[req.Token] = c
			return q, nil
		}
		q, list := nextQuestion(c, c.demoQnList, DEMO)
		c.demoQnList = list
		cmap[req.Token] = c
		return q, nil
	}
	// TOOD(pawan) - Check if time is up
	if len(c.qnList) == 0 {
		q := &interact.Question{Id: END, Totscore: c.score}
		c.logFile.Close()
		return q, nil
	}
	// This means its the first test question.
	if len(c.qnList) == len(qnList) {
		c.logFile.WriteString(fmt.Sprintf("%v test_start\n", UTCTime()))
		c.testStart = time.Now()
	}
	q, list := nextQuestion(c, c.qnList, TEST)
	c.qnList = list
	cmap[req.Token] = c
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

func isCorrectAnswer(resp *interact.Response) (int, int64, error) {
	for i, que := range quizInfo[resp.TestType] {
		if que.Id == resp.Qid {
			if reflect.DeepEqual(resp.Aid, que.Correct) {
				return i, CORRECT, nil
			} else {
				return i, WRONG, nil
			}
		}
	}
	return -1, -1, errors.New("No matching question")
}

func UTCTime() string {
	return time.Now().UTC().Format("2006/01/02 15:04:05 MST")
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

	idx, status.Status, err = isCorrectAnswer(resp)

	if len(resp.Aid) > 0 && resp.Aid[0] != "skip" {
		if status.Status == 1 {
			c.score += quizInfo[resp.TestType][idx].Positive
		} else {
			c.score -= quizInfo[resp.TestType][idx].Negative
		}
	} else {
		if len(resp.Aid) > 1 {
			glog.Error("Got extra optoins with SKIP")
		}
	}

	cmap[resp.Token] = c
	// We log only if its a actual test question.
	if resp.TestType == TEST {
		c.logFile.WriteString(fmt.Sprintf("%s response %s %s %.1f\n", UTCTime(),
			resp.Qid, strings.Join(resp.Aid, ","), c.score))
	}
	return &status, err
}

func (s *server) SendAnswer(ctx context.Context,
	resp *interact.Response) (*interact.Status, error) {
	if err := isValidSession(resp.Token, resp.Sid); err != nil {
		return &interact.Status{}, err
	}
	return sendAnswer(resp)
}

// TODO(ashwin) - Add authentication
func (s *server) StreamChan(stream interact.GruQuiz_StreamChanServer) error {

	var stat interact.ServerStatus
	var wg sync.WaitGroup

	msg, err := stream.Recv()
	if err != nil {
		glog.Error(err)
	}
	token := msg.Token
	c := cmap[token]

	wg.Add(1)
	go func() {
		for {
			stat.TimeLeft = time.Now().Sub(c.testStart).String()
			if time.Now().Sub(c.testStart) > time.Duration(DURATION) {
				// TODO(pawan) - Log this to candidate file.
				fmt.Println("End test based on time")
				stat.Status = "END"
			} else {
				stat.Status = "ONGOING"
			}

			if err := stream.Send(&stat); err != nil {
				glog.Error(err)
			}

			if stat.Status == "END" {
				wg.Done()
			}
			time.Sleep(5 * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					glog.Error(err)
				} else {
					break
				}
			}
			c.logFile.WriteString(fmt.Sprintf("%v ping %s\n",
				UTCTime(), msg.CurrQuestion))
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func runGrpcServer(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		glog.WithField("err", err).Fatalf("Error running quiz server")
		return
	}
	glog.WithField("address", ln.Addr()).Info("Server listening")

	s := grpc.NewServer()
	interact.RegisterGruQuizServer(s, &server{})
	if err = s.Serve(ln); err != nil {
		glog.Fatalf("While serving gRpc requests", err)
	}
	return
}

type Candidate struct {
	name     string
	email    string
	validity time.Time
	score    float32
	// List of test questions that have not been asked to the candidate yet.
	qnList     []string
	demoQnList []string
	logFile    *os.File
	testStart  time.Time
	// session id of currently active session.
	sid string
}

func parseCandidateInfo(file string) error {
	cmap = make(map[string]Candidate)
	format := "2006/01/02 (MST)"
	f, err := os.Open(*candFile)
	if err != nil {
		log.Fatal(err)
		return nil
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
			continue
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

func extractQids(testType string) []string {
	var list []string
	for _, q := range quizInfo[testType] {
		list = append(list, q.Id)
	}
	return list
}

func extractQuizInfo(file string) map[string][]Question {
	// TODO - Error out if the qid or aid is repeated.
	var info map[string][]Question
	b, err := ioutil.ReadFile(file)
	if err != nil {
		glog.WithField("err", err).Fatal("Error while reading quiz info file")
	}
	err = yaml.Unmarshal(b, &info)
	if err != nil {
		glog.WithField("err", err).Fatal("Error while unmarshalling into yaml")
	}
	return info
}

func main() {
	flag.Parse()
	quizInfo = extractQuizInfo(*quizFile)
	qnList = extractQids(TEST)
	demoQnList = extractQids(DEMO)
	parseCandidateInfo(*candFile)
	// TODO(pawan) - Read testStart timings for candidates.
	runGrpcServer(*port)
}
