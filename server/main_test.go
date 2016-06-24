package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/gru/server/interact"
)

func TestIsCorrectAnswer(t *testing.T) {
	t.Skip()
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}

	r := interact.Response{Qid: "demo-2", Aid: []string{"demo-2a"}}
	idx, status := isCorrectAnswer(&r)
	if idx != 1 {
		t.Errorf("Expected index %d, Got: %d", 1, idx)
	}
	if status != WRONG {
		t.Errorf("Expected status %d, Got: %d", WRONG, status)
	}

	r = interact.Response{Qid: "demo-2", Aid: []string{"demo-2c", "demo-2a"}}
	idx, status = isCorrectAnswer(&r)

	if idx != 1 {
		t.Errorf("Expected index %d, Got: %d", 1, idx)
	}
	if status != CORRECT {
		t.Errorf("Expected status %d, Got: %d", CORRECT, status)
	}
}

func TestNextQuestion(t *testing.T) {
	t.Skip()
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}

	c := Candidate{questions: questions[:]}
	cmap = make(map[string]Candidate)
	cmap["testtoken"] = c
	q, err := nextQuestion(c, "testtoken", DEMO)
	if err != nil {
		t.Errorf("Expected nil error. Got: %v", err)
	}
	c = cmap["testtoken"]
	if c.demoQnsAsked != 1 {
		t.Errorf("Expected demoQnsAsked to be %v. Got: %v", 1,
			c.demoQnsAsked)
	}
	if len(c.questions) != 2 {
		t.Errorf("Expected questions to have len %v. Got: %v", 2,
			len(c.questions))
	}
	if q.Id != "demo-1" {
		t.Errorf("Expected question with id: %v. Got: %v", "demo-1",
			q.Id)
	}

	q, err = nextQuestion(c, "testtoken", DEMO)
	if err != nil {
		t.Errorf("Expected nil error. Got: %v", err)
	}
	c = cmap["testtoken"]
	if c.demoQnsAsked != 2 {
		t.Errorf("Expected demoQnsAsked to be %v. Got: %v", 2,
			c.demoQnsAsked)
	}
	if len(c.questions) != 1 {
		t.Errorf("Expected questions to have len %v. Got: %v", 1,
			len(c.questions))
	}
	if q.Id != "demo-3" {
		t.Errorf("Expected question with id: %v. Got: %v", "demo-3",
			q.Id)
	}

	q, err = nextQuestion(c, "testtoken", TEST)
	if err != nil {
		t.Errorf("Expected nil error. Got: %v", err)
	}
	c = cmap["testtoken"]
	if c.demoQnsAsked != 2 {
		t.Errorf("Expected demoQnsAsked to be %v. Got: %v", 2,
			c.demoQnsAsked)
	}
	if len(c.questions) != 0 {
		t.Errorf("Expected questions to have len %v. Got: %v", 0,
			len(c.questions))
	}
	if q.Id != "test-2" {
		t.Errorf("Expected question with id: %v. Got: %v", "test-2",
			q.Id)
	}
}

func TestGetQuestion(t *testing.T) {
	t.Skip()
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}
	parseCandidateFile("cand_test.txt")
	c := cmap["abcd1234"]
	c.questions = make([]Question, len(questions))
	copy(c.questions, questions)
	cmap["abcd1234"] = c

	req := &interact.Req{Token: "abcd1234"}
	q1, err := getQuestion(req)
	if err != nil {
		t.Error(err)
	}
	if q1.Id == END {
		t.Errorf("Expected q.Id not to be %s", END)
	}
	q2, err := getQuestion(req)
	if q2.Id == q1.Id {
		t.Errorf("Expected %s to be different from %s", q2.Id, q1.Id)
	}

	q3, err := getQuestion(req)
	if q3.Id == q1.Id || q3.Id == q2.Id {
		t.Errorf("Expected %s to be different from %s and %s", q3.Id,
			q1.Id, q2.Id)
	}
	if len(cmap["abcd1234"].questions) != 0 {
		t.Errorf("Expected demo qn list to be empty. Got: len %d",
			len(cmap["abcd1234"].questions))
	}

	q, err := getQuestion(req)
	if err != nil {
		t.Error(err)
	}
	if q.Id != END {
		t.Errorf("Expected q.Id to be %s. Got: %s", END, q.Id)
	}
}

func TestCheckToken(t *testing.T) {
	t.Skip()
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7),
		testStart: time.Now().Add(-2 * time.Hour)}
	cmap = make(map[string]Candidate)
	cmap["test_token"] = c
	err := checkToken(c)
	if err == nil {
		t.Errorf("Expected non-nil error. Got: nil")
	}

	c.testStart = time.Now().Add(-1 * time.Minute)
	c.validity = time.Now().AddDate(0, 0, -1)
	cmap["test_token"] = c
	err = checkToken(c)
	if err == nil {
		t.Errorf("Expected non-nil error. Got: nil")
	}

	c.validity = time.Now().AddDate(0, 0, 7)
	cmap["test_token"] = c
	err = checkToken(c)
	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
}

func TestAuthenticate(t *testing.T) {
	t.Skip()
	tokenId := "test_token"
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7),
		questions: questions[:]}
	c.testStart = time.Now().Add(-2 * time.Minute)
	c.questions = questions[:1]
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
	if s.State != interact.Quiz_TEST_STARTED {
		t.Errorf("Expected state to be %d,Got: %d", s.State)
	}
	//TODO(pawan) - test other values fo State

	c.testStart = time.Now().Add(-2 * time.Hour)
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

	// Testing the case when log file doesn't exist
	tokenId = "test_token2"
	c = Candidate{email: "ashwin@dgraph.io", validity: time.Now().AddDate(0, 0, 7)}
	cmap[tokenId] = c
	token.Id = tokenId
	s, err = authenticate(&token)
	if err != nil {
		t.Errorf("Expected nil error. Got: %s", err.Error())
	}
	if _, err = os.Stat("logs/test_token2.log"); os.IsNotExist(err) {
		t.Error("Expected file to exist", err)
	}
	if s.Id == "" {
		t.Errorf("Expected non-empty sessionId. Got: %s", s.Id)
	}
	if err = os.Remove("logs/test_token2.log"); err != nil {
		t.Error(err)
	}
}

func TestLoadCandInfo(t *testing.T) {
	t.Skip()
	tokenId := "test_token"
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7)}
	cmap = make(map[string]Candidate)
	qnList = []string{"demo-1", "demo-2", "demo-3"}

	err := c.loadCandInfo(tokenId)
	if err != nil {
		t.Error(err)
	}
	c = cmap[tokenId]

	if c.score != 15.0 {
		t.Errorf("Expected score %f. Got: %f", 15.0, c.score)
	}
	// if !reflect.DeepEqual(c.qnList, []string{"demo-3"}) {
	// 	t.Error("Expected qn list doesn't match")
	// }
}

func TestSendAnswer(t *testing.T) {
	t.Skip()
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}
	r := interact.Response{Qid: "demo-2", Aid: []string{"demo-2a", "demo-2c"},
		Token: "test_token"}
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

func TestSliceDiff(t *testing.T) {
	// qnList := []string{"q7", "q9", "q1", "q2", "q5"}
	// qnsAsked := []string{"q5", "q1"}
	// qnsToAsk := sliceDiff(qnList, qnsAsked)

	// if len(qnsToAsk) != 3 {
	// 	t.Errorf("Expected slice to have len: %d. Got: %d", 3, len(qnsToAsk))
	// }
	// if !reflect.DeepEqual(qnsToAsk, []string{"q7", "q9", "q2"}) {
	// 	t.Error("qnsToAsk doesn't have all the qns.")
	// }
}

func TestCheckIds(t *testing.T) {
	t.Skip()
	qns := []Question{{Id: "qn1"}, {Id: "qn1"}}
	expectedError := "Qn Id has been used before: qn1"
	if err := checkIds(qns); err.Error() != expectedError {
		t.Errorf("Expected error to be %v. Got: %v", expectedError, err)
	}

	qns = []Question{
		{
			Id: "qn1",
			Opt: []Option{
				{Uid: "O1"},
				{Uid: "O2"},
			},
		},
		{
			Id: "qn2",
			Opt: []Option{
				{Uid: "O3"},
				{Uid: "O2"},
			},
		},
	}
	expectedError = "Ans Id has been used before: O2"
	if err := checkIds(qns); err.Error() != expectedError {
		t.Errorf("Expected error to be %v. Got: %v", expectedError, err)
	}

	qns = []Question{
		{
			Id: "qn1",
			Opt: []Option{
				{Uid: "O1"},
				{Uid: "O2"},
			},
		},
		{
			Id: "qn2",
			Opt: []Option{
				{Uid: "O3"},
				{Uid: "O4"},
			},
		},
	}
	if err := checkIds(qns); err != nil {
		t.Errorf("Expected error to be nil. Got: %v", err)
	}
}
