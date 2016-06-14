package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dgraph-io/gru/server/interact"
)

func TestIsCorrectAnswer(t *testing.T) {
	quizInfo = extractQuizInfo("demo_test.yaml")
	r := interact.Response{Qid: "demo-2", Aid: []string{"demo-2a"},
		TestType: DEMO}
	idx, status, err := isCorrectAnswer(&r)
	if err != nil {
		t.Error(err)
	}

	if idx != 1 {
		t.Errorf("Expected index %d, Got: %d", 1, idx)
	}
	if status != WRONG {
		t.Errorf("Expected status %d, Got: %d", WRONG, status)
	}

	r = interact.Response{Qid: "demo-2", Aid: []string{"demo-2c", "demo-2a"},
		TestType: DEMO}
	idx, status, err = isCorrectAnswer(&r)
	if err != nil {
		t.Error(err)
	}

	if idx != 1 {
		t.Errorf("Expected index %d, Got: %d", 1, idx)
	}
	if status != CORRECT {
		t.Errorf("Expected status %d, Got: %d", CORRECT, status)
	}
}

func TestNextQuestion(t *testing.T) {
	quizInfo = extractQuizInfo("demo_test.yaml")

	c := Candidate{}
	c.demoQnList = extractQids(DEMO)[:]

	idx, q := nextQuestion(c, DEMO)
	if idx >= 3 {
		t.Errorf("Expected idx to be less than %d, Got: %d", 3, idx)
	}
	if q == nil {
		t.Errorf("Expected qn got nil")
	}
}

func TestGetQuestion(t *testing.T) {
	quizInfo = extractQuizInfo("demo_test.yaml")
	demoQnList = extractQids(DEMO)
	parseCandidateInfo("cand_test.txt")
	c := cmap["abcd1234"]
	c.demoQnList = demoQnList[:]
	cmap["abcd1234"] = c

	req := &interact.Req{TestType: DEMO, Token: "abcd1234"}
	q, err := getQuestion(req)
	if err != nil {
		t.Error(err)
	}
	if q.Id == END {
		t.Errorf("Expected q.Id not to be %s", END)
	}
	getQuestion(req)
	getQuestion(req)
	q, err = getQuestion(req)
	if err != nil {
		t.Error(err)
	}
	if q.Id != END {
		t.Errorf("Expected q.Id to be %s. Got: %s", END, q.Id)
	}
}

func TestCheckToken(t *testing.T) {
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7),
		testStart: time.Now().Add(-2 * time.Hour)}
	cmap = make(map[string]Candidate)
	cmap["test_token"] = c
	valid, err := checkToken(c)

	if valid != false {
		t.Errorf("Expected valid to be %t. Got: %t", false, valid)
	}
	if err == nil {
		t.Errorf("Expected non-nil error. Got: nil")
	}

	c.testStart = time.Now().Add(-1 * time.Minute)
	c.validity = time.Now().AddDate(0, 0, -1)
	cmap["test_token"] = c
	valid, err = checkToken(c)
	if valid != false {
		t.Errorf("Expected valid to be %t. Got: %t", false, valid)
	}
	if err == nil {
		t.Errorf("Expected non-nil error. Got: nil")
	}

	c.validity = time.Now().AddDate(0, 0, 7)
	cmap["test_token"] = c
	valid, err = checkToken(c)
	if valid != true {
		t.Errorf("Expcted valid to be %t. Got: %t", true, valid)
	}
}

func TestAuthenticate(t *testing.T) {
	tokenId := "test_token"

	quizInfo = extractQuizInfo("demo_test.yaml")
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7),
		demoQnList: extractQids(DEMO)[:]}
	cmap = make(map[string]Candidate)
	cmap[tokenId] = c
	token := interact.Token{Id: tokenId}
	s, err := authenticate(&token)
	if err != nil {
		t.Errorf("Expected nil error. Got: %s", err.Error())
	}
	if s.Id == "" {
		t.Errorf("Expected non-empty sessionId. Got: %s", s.Id)
	}

	c.testStart = time.Now().Add(-2 * time.Hour)
	cmap = make(map[string]Candidate)
	cmap[tokenId] = c
	token = interact.Token{Id: tokenId}
	_, err = authenticate(&token)
	if err == nil {
		t.Errorf("Expected non-nil error. Got: %s", err.Error())
	}

	c.testStart = time.Now().Add(-1 * time.Minute)
	cmap[tokenId] = c
	s, err = authenticate(&token)
	if s.Id == "" {
		t.Errorf("Expected non-empty sessionId. Got: %s", s.Id)
	}
}

func TestLoadCandInfo(t *testing.T) {
	tokenId := "test_token"
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7)}
	cmap = make(map[string]Candidate)

	err := c.loadCandInfo(tokenId)
	if err != nil {
		t.Error(err)
	}
	c = cmap[tokenId]

	if c.score != 30.0 {
		t.Errorf("Expected score %f. Got: %f", 30.0, c.score)
	}
	if !reflect.DeepEqual(c.qnList, []string{"A", "B", "C"}) {
		t.Error("Expected qn list doesn't match")
	}
}

func TestSendAnswer(t *testing.T) {
	quizInfo = extractQuizInfo("demo_test.yaml")
	r := interact.Response{Qid: "demo-2", Aid: []string{"demo-2a", "demo-2c"},
		TestType: DEMO, Token: "test_token"}
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7),
		testStart: time.Now().Add(-2 * time.Minute)}
	f, err := ioutil.TempFile("", "test_token")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())

	c.logFile = f
	cmap = make(map[string]Candidate)
	cmap["test_token"] = c

	s, err := sendAnswer(&r)
	if err != nil {
		t.Error("Expected error to be nil.")
	}
	if s.Status != 1 {
		t.Errorf("Expected status to be 1. Got: %d", s.Status)
	}
	if cmap["test_token"].score <= 0.0 {
		t.Errorf("Expected positive score. Got: -%f", cmap["test_token"].score)
	}
	c.score = 0.0
	cmap["test_token"] = c

	r.Aid = []string{"demo-2b"}
	s, err = sendAnswer(&r)
	if err != nil {
		t.Error("Expected error to be nil.")
	}
	if s.Status != 2 {
		t.Errorf("Expected status to be 2. Got: %d", s.Status)
	}
	if cmap["test_token"].score > 0.0 {
		t.Errorf("Expected negative score. Got: %f", cmap["test_token"].score)
	}

	c.score = 0.0
	cmap["test_token"] = c
	r.Aid = []string{"skip"}
	s, err = sendAnswer(&r)
	if err != nil {
		t.Error("Expected error to be nil.")
	}
	// TODO(pawan) - Have another status code for skip
	if s.Status != 2 {
		t.Errorf("Expected status to be 0. Got: %d", s.Status)
	}
	if cmap["test_token"].score != 0.0 {
		t.Errorf("Expected 0.0 score. Got: %f", cmap["test_token"].score)
	}
}
