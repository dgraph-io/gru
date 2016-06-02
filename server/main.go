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

type server struct{}

func (s *server) SendQuestion(ctx context.Context,
	req *interact.Req) (*interact.Question, error) {

	var que *interact.Question
	var err error

	//	que = getNextQuestion()

	return que, err
}

func (s *server) SendAnswer(ctx context.Context,
	resp *interact.Response) (*interact.Status, error) {

	var status *interact.Status
	var err error

	fmt.Println(resp.Aid)

	status.Status, err = isCorrectAnswer(resp.Qid, resp.Aid)

	// Check for end of test and change status
	return status, err
}

func isCorrectAnswer(qid string, opts []string) (int64, error) {

	for _, que := range quizInfo[candidateInfo["tname"]] {
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

/*
func getNextQuestion() *interact.Question {}
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

type Option struct {
	Uid string
	Str string
}

type T struct {
	Qid      string
	Question string
	Correct  []string
	Opt      []map[string]string
	Score    int
	Tag      string
}

type candidate struct {
	token          string
	name           string
	candidateEmail string
	valid          time.Duration
	invitorEmail   string
}

type session struct {
	id                string
	questionSeq       []string
	attemptedQuestion map[string]string
	totScore          float64
}

var (
	quizFile      = flag.String("quiz", "", "Input question file")
	port          = flag.String("port", ":8888", "Port on which server listens")
	candFile      = flag.String("cand", "", "Candidate inforamation file")
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

		cand := &candidate{}

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

	parseCandidateInfo(*candFile)

	err := yaml.Unmarshal(data, &quizInfo)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", quizInfo["test"])

	_, err = yaml.Marshal(&quizInfo)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	//fmt.Printf("--- m dump:\n%s\n\n", string(d))

	runGrpcServer(*port)
}
