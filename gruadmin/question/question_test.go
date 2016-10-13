package question

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func startDgraph(t *testing.T) (string, string) {
	posting, err := ioutil.TempDir("", "posting")
	if err != nil {
		t.Fatal(err)
	}

	mutations, err := ioutil.TempDir("", "mutations")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("./dgraph", "--cluster", "1:localhost:12345", "--m",
		mutations, "--p", posting, "&")
	// out, _ := cmd.CombinedOutput()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	//fmt.Println(string(out))
	// Waiting for instance to become leader so that it can assign uids.
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
	p, m := startDgraph(t)
	defer stopDgraph(t, p, m)
	body := `{ not even json }`
	req := httptest.NewRequest("PUT", "http://example.com/foo", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Edit(w, req)

	body = `{
    "name": "Validation errors",
    "text": "This is question text"
}`
	req = httptest.NewRequest("PUT", "http://example.com/foo", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	Edit(w, req)

	adminRouter := mux.NewRouter().PathPrefix("/admin").Subrouter().StrictSlash(true)
	adminRouter.HandleFunc("/question/{id}", Edit).Methods("PUT", "OPTIONS")

	body = `{
    "_uid_": "0x1b2b484c8b97e157",
	"name": "Question 1",
	"text": "Question test",
    "positive": 22,
    "negative": 11,
    "tags": [{"_uid_": "0x603606c14a73cb25", "name": "anotherone"},{ "_uid_": "0x9abf176412f8207e",
          "name": "again", "is_delete": true }],
    "options": [{"_uid_": "0x15a481bb495278b1","name": "option 1", "is_correct": true}, {"_uid_": "0xedfca6109fbcdbf1","name": "option 2", "is_correct": true}, {"_uid_": "0x7d702df457868258","name": "option 2", "is_correct": false}]
}`
	req = httptest.NewRequest("PUT", "http://example.com/admin/question/0x1b2b484c8b97e157", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	adminRouter.ServeHTTP(w, req)
	require.Contains(t, w.Body.String(), "Question updated successfully.")
}

func TestIndex(t *testing.T) {
	p, m := startDgraph(t)
	defer stopDgraph(t, p, m)

	// Lets add something first.
	body := `{
    "name": "Question 1",
    "text": "This is question text",
    "positive": 5.0,
    "negative": 2.5,
    "tags": [{"uid": "0xe5579130a965f0e7", "name": "tag 2"},{"uid": "", "name": "tag 3"}],
    "options": [{"name": "option 1", "is_correct": true}, {"name": "option 2", "is_correct": true}]
}`
	req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Add(w, req)

	req = httptest.NewRequest("POST", "http://localhost:8082/admin/get-all-questions", bytes.NewBufferString(`{"id": ""}`))
	w = httptest.NewRecorder()
	Index(w, req)
	require.Contains(t, w.Body.String(), "This is question text")
}

func TestGet(t *testing.T) {
	p, m := startDgraph(t)
	defer stopDgraph(t, p, m)

	// Lets add something first.
	body := `{
    "name": "Question 1",
    "text": "This is question text",
    "positive": 5.0,
    "negative": 2.5,
    "tags": [{"uid": "0xe5579130a965f0e7", "name": "tag 2"},{"uid": "", "name": "tag 3"}],
    "options": [{"name": "option 1", "is_correct": true}, {"name": "option 2", "is_correct": true}]
}`
	req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Add(w, req)

	adminRouter := mux.NewRouter().PathPrefix("/admin").Subrouter().StrictSlash(true)
	adminRouter.HandleFunc("/question/{id}", Get).Methods("GET", "OPTIONS")
	req = httptest.NewRequest("GET", "http://localhost:8082/admin/question/", nil)
	w = httptest.NewRecorder()
	adminRouter.ServeHTTP(w, req)
	// TODO - Test an actual complete get.
}
