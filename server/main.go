// sample run : ./server --cand testCand.csv --quiz testYML

package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"time"

	"google.golang.org/grpc"

	"github.com/dgraph-io/dgraph/x"
	"github.com/dgraph-io/gru/server/interact"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

var data string
var glog = x.Log("Gru Server")
var totScore float32
var qList []string

type server struct{}

func (s *server) GetQuestion(ctx context.Context,
	req *interact.Req) (*interact.Question, error) {

	var que *interact.Question
	var err error

	if len(qList) == len(quizInfo["test"]) {
		que := &interact.Question{
			Qid:        "END",
			Question:   "",
			Options:    nil,
			IsMultiple: false,
			Positive:   0,
			Negative:   0,
			Totscore:   totScore,
		}
		return que, nil
	}

	que = getNextQuestion()
	qList = append(qList, que.Qid)
	fmt.Println(qList)

	return que, err
}

func (s *server) SendAnswer(ctx context.Context,
	resp *interact.Response) (*interact.Status, error) {

	var status interact.Status
	var err error
	var idx int

	fmt.Println(resp.Aid)

	status.Status, err = isCorrectAnswer(resp.Qid, resp.Aid, resp.Token)

	for i, que := range quizInfo["test"] {
		if que.Qid == resp.Qid {
			idx = i
		}
	}

	if len(resp.Aid) > 0 && resp.Aid[0] != "skip" {
		if status.Status == 1 {
			totScore += quizInfo["test"][idx].Score
		} else {
			totScore -= quizInfo["test"][idx].Score
		}
	} else {
		if len(resp.Aid) > 1 {
			glog.Error("Got extra optoins with SKIP")
		}
	}

	fmt.Println(status.Status, totScore)

	// Check for end of test and change status
	return &status, err
}

func isCorrectAnswer(qid string, opts []string, token string) (int64, error) {

	for _, que := range quizInfo["test"] {
		if que.Qid == qid {
			if reflect.DeepEqual(opts, que.Correct) {
				return 1, nil
			} else {
				return 2, nil
			}
		}
	}

	return -1, errors.New("No matching question")
}

func getNextQuestion() *interact.Question {

	idx := rand.Intn(3)

	var opts []*interact.Answer

	for _, mp := range quizInfo["test"][idx].Opt {
		it := &interact.Answer{
			Id:  mp["uid"],
			Ans: mp["str"],
		}

		opts = append(opts, it)
	}

	var isM bool

	if len(quizInfo["test"][idx].Correct) > 1 {
		isM = true
	}

	que := &interact.Question{
		Qid:        quizInfo["test"][idx].Qid,
		Question:   quizInfo["test"][idx].Question,
		Options:    opts,
		IsMultiple: isM,
		Positive:   quizInfo["test"][idx].Score,
		Negative:   quizInfo["test"][idx].Score,
		Totscore:   totScore,
	}
	return que
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

type Option struct {
	Uid string
	Str string
}

type T struct {
	Qid      string
	Question string
	Correct  []string
	Opt      []map[string]string
	Score    float32
	Tag      string
}

type candidate struct {
	token          string
	name           string
	candidateEmail string
	valid          time.Duration
	invitorEmail   string
	tname          string
}

type session struct {
	id                string
	questionSeq       []string
	attemptedQuestion map[string]string
	totScore          float32
}

var (
	quizFile      = flag.String("quiz", "testYML", "Input question file")
	port          = flag.String("port", ":8888", "Port on which server listens")
	candFile      = flag.String("cand", "testCand.csv", "Candidate inforamation file")
	quizInfo      map[interface{}][]T
	candidateInfo map[string]*candidate
	sessionInfo   map[string]*session
)

func parseCandidateInfo(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	csvr := csv.NewReader(f)

	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}

		cand := &candidate{
			token:          row[2],
			name:           row[0],
			candidateEmail: row[1],
			valid:          time.Duration(10 * time.Second),
			tname:          row[3],
		}
		candidateInfo[row[2]] = cand
		fmt.Println(row, cand)
		// store info in struct

	}
	return nil
}

func main() {
	flag.Parse()

	buf := bytes.NewBuffer(nil)
	f, _ := os.Open(*quizFile)
	io.Copy(buf, f)
	data := buf.Bytes()

	candidateInfo = make(map[string]*candidate)
	parseCandidateInfo(*candFile)

	err := yaml.Unmarshal(data, &quizInfo)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	/*
		fmt.Printf("--- m:\n%v\n\n", quizInfo["test"])

		_, err = yaml.Marshal(&quizInfo)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		//fmt.Printf("--- m dump:\n%s\n\n", string(d))
	*/
	fmt.Println(candidateInfo)
	runGrpcServer(*port)
}
