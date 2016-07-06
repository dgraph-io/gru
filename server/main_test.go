package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/gru/server/interact"
)

func TestIsCorrectAnswer(t *testing.T) {
	maxDemoQns = 3
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}

	r := interact.Response{Qid: "demo-2", Aid: []string{"demo-2a"}}
	idx, score := isCorrectAnswer(&r)
	if idx != 2 {
		t.Errorf("Expected index %d, Got: %d", 2, idx)
	}
	if score != 5.0 {
		t.Errorf("Expected score %f, Got: %f", 5.0, score)
	}

	r = interact.Response{Qid: "demo-2", Aid: []string{"demo-2c", "demo-2a"}}
	idx, score = isCorrectAnswer(&r)
	if score != 10.0 {
		t.Errorf("Expected score %f, Got: %f", 10.0, score)
	}

	r = interact.Response{Qid: "demo-2", Aid: []string{"demo-2d", "demo-2b"}}
	idx, score = isCorrectAnswer(&r)
	if score != -10.0 {
		t.Errorf("Expected score %f, Got: %f", -10.0, score)
	}

	//Single choice questions
	r = interact.Response{Qid: "demo-1", Aid: []string{"demo-1b"}}
	idx, score = isCorrectAnswer(&r)
	if score != 5 {
		t.Errorf("Expected score %f, Got: %f", 5.0, score)
	}

	r = interact.Response{Qid: "demo-1", Aid: []string{"demo-1c"}}
	idx, score = isCorrectAnswer(&r)
	if score != -2.5 {
		t.Errorf("Expected score %f, Got: %f", -2.5, score)
	}
}

func TestNextQuestion(t *testing.T) {
	maxDemoQns = 3
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
	if len(c.questions) != 3 {
		t.Errorf("Expected questions to have len %v. Got: %v", 3,
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
	if len(c.questions) != 2 {
		t.Errorf("Expected questions to have len %v. Got: %v", 2,
			len(c.questions))
	}
	if q.Id != "demo-2" {
		t.Errorf("Expected question with id: %v. Got: %v", "demo-2",
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
	if len(c.questions) != 1 {
		t.Errorf("Expected questions to have len %v. Got: %v", 1,
			len(c.questions))
	}
	if q.Id != "test-2" {
		t.Errorf("Expected question with id: %v. Got: %v", "test-2",
			q.Id)
	}
}

func TestGetQuestion(t *testing.T) {
	maxDemoQns = 3
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}
	parseCandidateFile("cand_test.txt")
	c := cmap["abcd1234"]
	c.questions = make([]Question, len(questions))
	c.logFile, err = ioutil.TempFile("", "gru")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(c.logFile.Name())
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
	if q3.Id != "demo-3" {
		t.Errorf("Expected qn Id to be %v. Got: %v", "demo-3", q3.Id)
	}

	q4, err := getQuestion(req)
	if q4.Id != "DEMOEND" {
		t.Errorf("Expected qn Id to be %v. Got: %v", "DEMOEND", q4.Id)
	}

	q5, err := getQuestion(req)
	if q5.Id != "test-2" {
		t.Errorf("Expected qn Id to be %v. Got: %v", "test-2", q5.Id)
	}

	if len(cmap["abcd1234"].questions) != 0 {
		t.Errorf("Expected qn list to be empty. Got: len %d",
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
	maxDemoQns = 3
	tokenId := "test_token"
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7),
		questions: questions[:]}
	c.logFile, err = ioutil.TempFile("", "gru")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(c.logFile.Name())
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
	if s.State != interact.Quiz_TEST_NOT_TAKEN {
		t.Errorf("Expected state to be %d,Got: %d",
			interact.Quiz_TEST_NOT_TAKEN, s.State)
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
	tokenId := "test_token"
	c := Candidate{email: "pawan@dgraph.io", validity: time.Now().AddDate(0, 0, 7)}
	cmap = make(map[string]Candidate)
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}

	err = c.loadCandInfo(tokenId)
	if err != nil {
		t.Error(err)
	}
	if c.score != 10.0 {
		t.Errorf("Expected score %f. Got: %f", 10.0, c.score)
	}
}

func TestSendAnswer(t *testing.T) {
	maxDemoQns = 3
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

	_, err = sendAnswer(&r)
	if err != nil {
		t.Error("Expected error to be nil.")
	}
	if cmap["test_token"].score <= 0.0 {
		t.Errorf("Expected positive score. Got: -%f", cmap["test_token"].score)
	}
	c.score = 0.0
	cmap["test_token"] = c

	r.Aid = []string{"demo-2b"}
	_, err = sendAnswer(&r)
	if err != nil {
		t.Error("Expected error to be nil.")
	}
	if cmap["test_token"].score > 0.0 {
		t.Errorf("Expected negative score. Got: %f", cmap["test_token"].score)
	}

	c.score = 0.0
	cmap["test_token"] = c
	r.Aid = []string{"skip"}
	_, err = sendAnswer(&r)
	if err != nil {
		t.Error("Expected error to be nil.")
	}
	if cmap["test_token"].score != 0.0 {
		t.Errorf("Expected 0.0 score. Got: %f", cmap["test_token"].score)
	}
}

func TestSliceDiff(t *testing.T) {
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Error(err)
	}
	qnsAsked := []string{"demo-2"}
	qnsToAsk := sliceDiff(questions, qnsAsked)

	if len(qnsToAsk) != 3 {
		t.Errorf("Expected slice to have len: %d. Got: %d", 3, len(qnsToAsk))
	}
}

func TestCheckTest(t *testing.T) {
	qns := []Question{
		{
			Id: "qn1",
			Opt: []Option{
				{Uid: "O1"},
			},
			Correct: []string{"O1"},
		},
		{
			Id: "qn1",
			Opt: []Option{
				{Uid: "O3"},
			},
			Correct: []string{"O3"},
		},
	}
	expectedError := "Id has been used before: qn1"
	if err := checkTest(qns); err.Error() != expectedError {
		t.Errorf("Expected error to be %v. Got: %v", expectedError, err)
	}

	qns = []Question{
		{
			Id: "qn1",
			Opt: []Option{
				{Uid: "O1"},
				{Uid: "O2"},
			},
			Correct: []string{"O1"},
		},
		{
			Id: "qn2",
			Opt: []Option{
				{Uid: "O3"},
				{Uid: "O2"},
			},
			Correct: []string{"O3"},
		},
	}
	expectedError = "Id has been used before: O2"
	if err := checkTest(qns); err.Error() != expectedError {
		t.Errorf("Expected error to be %v. Got: %v", expectedError, err)
	}

	qns = []Question{{Id: "qn1", Tags: []string{"Demo"}}}
	expectedError = "Tag: Demo for qn: qn1 should start with a lowercase character"
	if err := checkTest(qns); err.Error() != expectedError {
		t.Errorf("Expected error to be %v. Got: %v", expectedError, err)
	}

	qns = []Question{
		{
			Id: "qn1",
			Opt: []Option{
				{Uid: "O1"},
				{Uid: "O2"},
			},
			Correct:  []string{"O2", "O1"},
			Positive: 2.5,
			Negative: 1,
			Tags:     []string{"demo"},
		},
	}
	expectedError = "Negative score less than positive for multi-choice qn: qn1"
	if err := checkTest(qns); err.Error() != expectedError {
		t.Errorf("Expected error to be %v. Got: %v", expectedError, err)
	}

	maxDemoQns = 3
	var err error
	questions, err = extractQuizInfo("demo_test.yaml")
	if err != nil {
		t.Errorf("Expected error to be nil. Got: %v", err)
	}
}

func TestCandfileRead(t *testing.T) {
	cmap = make(map[string]Candidate)
	questions, err := extractQuizInfo("demo_test.yaml")
	candFile, err := ioutil.TempFile("", "candFile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(candFile.Name())

	content := []byte("Mallory a test-mail@gmail.com 2017/12/06 IST wxwwr43e332\n")
	if _, err := candFile.Write(content); err != nil {
		t.Fatal(err)
	}
	parseCandidateFile(candFile.Name())
	c := cmap["wxwwr43e332"]
	c.questions = make([]Question, len(questions))
	c.logFile, err = ioutil.TempFile("", "gru")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(c.logFile.Name())
	copy(c.questions, questions)
	cmap["wxwwr43e332"] = c

	req := &interact.Req{Token: "wxwwr43e332"}
	_, err = getQuestion(req)
	if err != nil {
		t.Error(err)
	}

	content = []byte("Mary a test-mail@gmail.com 2017/12/06 IST fefvevrev3e332\n")
	if _, err := candFile.Write(content); err != nil {
		t.Fatal(err)
	}
	parseCandidateFile(candFile.Name())

	if len(questions) == len(cmap["wxwwr43e332"].questions) {
		t.Fatal("Candidate object updated on reading Candidate file")
	}

	if _, ok := cmap["fefvevrev3e332"]; !ok {
		t.Fatal("New candidate info not added")
	}

	if cmap["fefvevrev3e332"].email != "test-mail@gmail.com" {
		t.Fatal("Candidate info incorrect")
	}

	if err := candFile.Close(); err != nil {
		t.Fatal(err)
	}

}
