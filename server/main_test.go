package main

import (
	"fmt"
	"os"
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
	cand, valid := checkToken("test_token")

	if cand.email != "pawan@dgraph.io" {
		t.Errorf("Expected candidate email to be %s.Got: %s", "pawan@draph.io",
			cand.email)
	}
	if valid != false {
		t.Errorf("Expcted valid to be %t. Got: %t", false, valid)
	}

	c.testStart = time.Now().Add(-1 * time.Minute)
	c.validity = time.Now().AddDate(0, 0, -1)
	cmap["test_token"] = c
	cand, valid = checkToken("test_token")
	if cand.email != "pawan@dgraph.io" {
		t.Errorf("Expected candidate email to be %s.Got: %s", "pawan@draph.io",
			cand.email)
	}
	if valid != false {
		t.Errorf("Expcted valid to be %t. Got: %t", false, valid)
	}

	c.validity = time.Now().AddDate(0, 0, 7)
	cmap["test_token"] = c
	cand, valid = checkToken("test_token")
	if cand.email != "pawan@dgraph.io" {
		t.Errorf("Expected candidate email to be %s.Got: %s", "pawan@draph.io",
			cand.email)
	}
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

	// TODO(pawan) - Check auth token and sessionId is written to file.
	err = os.Remove(fmt.Sprintf("logs/%s.log", tokenId))
	if err != nil {
		t.Error(err)
	}
}
