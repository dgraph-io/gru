package quiz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/server"
)

type Quiz struct {
	Name       string
	Duration   string
	Start_Date string
	End_Date   string
	Questions  []string
}

type EditQuiz struct {
	Id         string
	Name       string
	Duration   string
	Start_Date string
	End_Date   string
	Questions  []string
}

func Add(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var quiz Quiz

	err := json.NewDecoder(r.Body).Decode(&quiz)
	if err != nil {
		panic(err)
	}
	//question_ids = quiz.Questions
	quiz_mutation := "mutation { set { <rootQuiz> <quiz> <_new_:quiz> . \n	<_new_:quiz> <name> \"" + quiz.Name +
		"\" . \n <_new_:quiz> <duration> \"" + quiz.Duration +
		"\" . \n <_new_:quiz> <start_date> \"" + quiz.Start_Date +
		"\" . \n <_new_:quiz> <end_date> \"" + quiz.End_Date + "\" . \n"
	for i := 0; i < len(quiz.Questions); i++ {
		quiz_mutation += "<_new_:quiz> <quiz.question> <_uid_:" + quiz.Questions[i] + "> .\n"
	}
	quiz_mutation += " }}"
	fmt.Println(quiz_mutation)
	_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
	if err != nil {
		panic(err)
	}
	stats := &server.Response{true, "Quiz Successfully Saved!"}
	quiz_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(quiz_json_response)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	var quiz EditQuiz

	err := json.NewDecoder(r.Body).Decode(&quiz)
	if err != nil {
		panic(err)
	}
	//question_ids = quiz.Questions
	quiz_mutation := "mutation { set { <_uid_:" + quiz.Id + "> <name> \"" + quiz.Name +
		"\" . \n <_uid_:" + quiz.Id + "> <duration> \"" + quiz.Duration +
		"\" . \n <_uid_:" + quiz.Id + "> <start_date> \"" + quiz.Start_Date +
		"\" . \n <_uid_:" + quiz.Id + "> <end_date> \"" + quiz.End_Date + "\" . \n"
	for i := 0; i < len(quiz.Questions); i++ {
		// question_id := strconv.FormatUint(quiz.Questions[i], 10)
		quiz_mutation = quiz_mutation + "<_uid_:" + quiz.Id + "> <quiz.question> <_uid_:" + quiz.Questions[i] + "> .\n"
	}
	quiz_mutation += " }}"
	fmt.Println(quiz_mutation)
	_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
	if err != nil {
		panic(err)
	}
	stats := &server.Response{true, "Quiz Successfully Updated!"}
	quiz_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(quiz_json_response)
}

func Index(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	quiz_mutation := "{debug(_xid_: rootQuiz) { quiz { _uid_ name duration quiz.question { text }  }  }}"
	quiz_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
	if err != nil {
		panic(err)
	}
	defer quiz_response.Body.Close()
	quiz_body, err := ioutil.ReadAll(quiz_response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(quiz_body))

	jsonResp, err := json.Marshal(string(quiz_body))
	if err != nil {
		panic(err)
	}

	w.Write(jsonResp)
	w.Header().Set("Content-Type", "application/json")
}
