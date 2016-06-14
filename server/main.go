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

func checkToken(c Candidate) (bool, error) {

	if time.Now().UTC().After(c.validity) {
		return false, errors.New("Your token has expired.")
	}
	// Initially testStart is zero, but after candidate has taken the
	// test once, it shouldn't be zero.
	if !c.testStart.IsZero() && time.Now().UTC().After(c.testStart.Add(DURATION)) {
		return false,
			errors.New(fmt.Sprintf("%v since you started the test for the first time are already over.",
				DURATION))
	}
	return true, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (c Candidate) loadCandInfo(token string) error {
	f, err := os.Open(fmt.Sprintf("logs/%s.log", token))
	if err != nil {
		return err
	}

	var score float32
	testQids := []string{}
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
			if len(testQids) > 0 && splits[4] == testQids[len(testQids)-1] {
				continue
			}
			testQids = append(testQids, splits[4])
		case "ping":
			if len(testQids) > 0 && splits[4] == testQids[len(testQids)-1] {
				continue
			}
			testQids = append(testQids, splits[4])
		}
	}
	c.score = score
	c.logFile = f
	c.demoQnList = demoQnList
	c.qnList = testQids
	cmap[token] = c
	return nil
}

func authenticate(t *interact.Token) (*interact.Session, error) {
	var c Candidate
	var ok bool

	// This indicates there is no entry in the candidate file with this token.
	if c, ok = cmap[t.Id]; !ok {
		return nil, errors.New("Invalid token.")
	}

	if _, err := os.Stat(fmt.Sprintf("logs/%s.log", t.Id)); err == nil {
		c.loadCandInfo(t.Id)
	} else {
		f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", t.Id),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		c.logFile = f
		c.qnList = qnList[:]
		c.demoQnList = demoQnList[:]
	}

	if ok, err := checkToken(c); !ok {
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

func nextQuestion(c Candidate, testType string) (int, *interact.Question) {
	var list []string
	if testType == DEMO {
		list = c.demoQnList
	} else {
		list = c.qnList
	}

	idx := rand.Intn(len(list))
	q := quizInfo[testType][idx]

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
	return idx, que
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
		idx, q := nextQuestion(c, testType)
		c.demoQnList = append(c.demoQnList[:idx], c.demoQnList[idx+1:]...)
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
	idx, q := nextQuestion(c, testType)
	c.qnList = append(c.qnList[:idx], c.qnList[idx+1:]...)
	cmap[req.Token] = c
	return q, nil
}

func (s *server) GetQuestion(ctx context.Context,
	req *interact.Req) (*interact.Question, error) {
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
	return sendAnswer(resp)
}

func (s *server) StreamChan(stream interact.GruQuiz_StreamChanServer) error {

	var stat interact.ServerStatus
	var wg sync.WaitGroup

	msg, err := stream.Recv()
	if err != nil {
		glog.Error(err)
	}
	token := msg.Token

	wg.Add(1)
	go func() {
		for {
			stat.TimeLeft = time.Now().Sub(cmap[token].testStart).String()
			if time.Now().Sub(cmap[token].testStart) > time.Duration(10*time.Second) {
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
			log.SetOutput(cmap[token].logFile)
			log.Println("ping", msg.CurrQuestion)
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

/*
func (s *server) StreamChanClient(stream interact.GruQuiz_StreamChanClientClient) error {


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
				log.Println("ping", msg.CurrQuestion)
			}
		}()

	return nil
}
*/
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
