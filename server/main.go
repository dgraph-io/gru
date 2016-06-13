// sample run : ./server --cand testCand.csv --quiz testYML

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/dgraph-io/dgraph/x"
	"github.com/dgraph-io/gru/server/interact"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

const (
	DEMO = "demo"
	TEST = "test"
	END  = "END"
)

var (
	quizFile = flag.String("quiz", "test.yml", "Input question file")
	port     = flag.String("port", ":8888", "Port on which server listens")
	candFile = flag.String("cand", "testCand.txt", "Candidate inforamation file")
	quizInfo map[interface{}][]Question
	cmap     map[string]Candidate
	glog     = x.Log("Gru Server")
	// List of question ids.
	qnList     []string
	demoQnList []string
)

type server struct{}

func checkToken(token string) (Candidate, bool) {
	if c, ok := cmap[token]; ok && time.Now().Before(c.validity) {
		return c, true
	}
	return Candidate{}, false
}

func (s *server) Authenticate(ctx context.Context,
	t *interact.Token) (*interact.Session, error) {
	var c Candidate
	var valid bool

	if c, valid = checkToken(t.Id); !valid {
		return nil, errors.New("Invalid token")
	}
	c.qnList = qnList[:]
	c.demoQnList = demoQnList[:]
	f, err := os.OpenFile(fmt.Sprintf("logs/cand-%s.log", t.Id),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	c.logFile = f
	cmap[t.Id] = c
	// TODO(pawan) - Generate session id and send that.
	return &interact.Session{Id: "abc"}, nil
}

type Option struct {
	Uid string
	Str string
}

type Question struct {
	Id       string
	Str      string
	Correct  []string
	Opt      []map[string]string
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

	for _, mp := range q.Opt {
		it := &interact.Answer{
			Id:  mp["uid"],
			Ans: mp["str"],
		}
		opts = append(opts, it)
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

func (s *server) GetQuestion(ctx context.Context,
	req *interact.Req) (*interact.Question, error) {
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
	idx, q := nextQuestion(c, testType)
	c.qnList = append(c.qnList[:idx], c.qnList[idx+1:]...)
	cmap[req.Token] = c
	return q, nil
}

func isCorrectAnswer(resp *interact.Response) (int, int64, error) {
	for i, que := range quizInfo[resp.TestType] {
		if que.Id == resp.Qid {
			if reflect.DeepEqual(resp.Aid, que.Correct) {
				return i, 1, nil
			} else {
				return i, 2, nil
			}
		}
	}
	return -1, -1, errors.New("No matching question")
}

func (s *server) SendAnswer(ctx context.Context,
	resp *interact.Response) (*interact.Status, error) {
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
		log.SetOutput(c.logFile)
		log.Println(resp.Qid, resp.Aid, status.Status, c.score)
	}

	return &status, err
}

func (s *server) StreamChan(stream interact.GruQuiz_StreamChanServer) error {

	stat := &interact.ServerStatus{
		"10",
	}
	if err := stream.Send(stat); err != nil {
		return err
	}
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

func main() {
	flag.Parse()
	buf := bytes.NewBuffer(nil)
	f, _ := os.Open(*quizFile)
	io.Copy(buf, f)
	data := buf.Bytes()
	err := yaml.Unmarshal(data, &quizInfo)
	if err != nil {
		glog.Fatalf("error: %v", err)
	}
	qnList = extractQids(TEST)
	demoQnList = extractQids(DEMO)
	parseCandidateInfo(*candFile)
	runGrpcServer(*port)
}
