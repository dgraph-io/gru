package question

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func startDgraph(t *testing.T) (string, string) {
	posting, err := ioutil.TempDir("", "posting")
	if err != nil {
		log.Fatal(err)
	}

	mutations, err := ioutil.TempDir("", "mutations")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("./dgraph", "--cluster", "1:localhost:12345", "--m",
		mutations, "--p", posting, "&")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	// Waiting for instance to become leader.
	time.Sleep(5 * time.Second)
	return posting, mutations
}

func stopDgraph(t *testing.T, p string, m string) {
	cmd := exec.Command("killall", "dgraph")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(m)
	os.RemoveAll(p)
}

func TestAdd(t *testing.T) {
	p, m := startDgraph(t)
	defer stopDgraph(t, p, m)
	body := `{ not even json }`
	req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Add(w, req)

	body = `{
    "name": "Validation errors",
    "text": "This is question text"
}`
	req = httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	Add(w, req)

	body = `{
    "name": "Question 1",
    "text": "This is question text",
    "positive": 5.0,
    "negative": 2.5,
    "tags": [{"uid": "0xe5579130a965f0e7", "name": "tag 2"},{"uid": "", "name": "tag 3"}],
    "options": [{"name": "option 1", "is_correct": true}, {"name": "option 2", "is_correct": false}]
}`
	req = httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	Add(w, req)
	require.Contains(t, w.Body.String(), "Question Successfully Saved")

	// Multiple correct options.
	body = `{
    "name": "Question 1",
    "text": "This is question text",
    "positive": 5.0,
    "negative": 2.5,
    "tags": [{"uid": "0xe5579130a965f0e7", "name": "tag 2"},{"uid": "", "name": "tag 3"}],
    "options": [{"name": "option 1", "is_correct": true}, {"name": "option 2", "is_correct": true}]
}`
	req = httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	Add(w, req)
	require.Contains(t, w.Body.String(), "Question Successfully Saved")
}

func TestEdit(t *testing.T) {
}
