package main

import (
	"testing"

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
