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
