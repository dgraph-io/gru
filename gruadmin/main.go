/*
 * Copyright 2016 DGraph Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/gruadmin/candidate"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/gorilla/mux"
)

var (
	port = flag.String("port", ":8082", "Port on which server listens")
)

type Tag struct {
	Uid       string `json:"_uid_"`
	Name      string
	Is_delete bool
}

type Option struct {
	Uid        string `json:"_uid_"`
	Name       string
	Is_correct bool
}

type Question struct {
	Uid      string `json:"_uid_"`
	Text     string
	Positive float64
	Negative float64
	Tags     []Tag
	Options  []Option
}

type TagFilter struct {
	UID string
}

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

type GetQuestion struct {
	Id string
}

type Status struct { // HTTP Response status
	Success bool
	Error   bool
	Message string
}

type QuestionAPIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	UIDS    struct {
		Question uint64 `json:question`
	}
}

type TagAPIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	UIDS    struct {
		Tag uint64 `json:tag`
	}
}

type QuizAPIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	UIDS    struct {
		Quiz uint64 `json:quiz`
	}
}

// API for "Adding Question" to Database
func AddQuestionHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	var ques Question

	if r.Method == "OPTIONS" {
		return
	}

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}

	fmt.Println(ques)

	// // Creating query mutation to save question informations.
	question_info_mutation := "mutation { set { <rootQuestion> <question> <_new_:question> . \n	<_new_:question> <text> \"" + ques.Text +
		"\" . \n <_new_:question> <positive> \"" + strconv.FormatFloat(ques.Positive, 'g', -1, 64) +
		"\" . \n <_new_:question> <negative> \"" + strconv.FormatFloat(ques.Negative, 'g', -1, 64) + "\" . \n "

	// Create and associate Options
	for l := range ques.Options {
		index := strconv.Itoa(l)
		question_info_mutation += "<_new_:option" + index + "> <name> \"" + ques.Options[l].Name +
			"\" .\n <_new_:question> <question.option> <_new_:option" + index + "> . \n "

		// If this option is correct answer
		if ques.Options[l].Is_correct == true {
			question_info_mutation += "<_new_:question> <question.correct> <_new_:option" + index + "> . \n "
		}
	}
	fmt.Println(ques.Tags)
	// Create and associate Tags
	for i := range ques.Tags {
		if ques.Tags[i].Uid != "" {
			fmt.Println(ques.Tags[i].Uid)
			question_info_mutation += "<_new_:question> <question.tag> <_uid_:" + ques.Tags[i].Uid +
				"> . \n <_uid_:" + ques.Tags[i].Uid + "> <tag.question> <_new_:question> . \n "
		} else {
			index := strconv.Itoa(i)
			question_info_mutation += "<_new_:tag" + index + "> <name> \"" + ques.Tags[i].Name +
				"\" .\n <_new_:question> <question.tag> <_new_:tag" + index +
				"> . \n <_new_:tag" + index + "> <tag.question> <_new_:question> . \n "
		}
	}

	question_info_mutation += " }}"
	fmt.Println(question_info_mutation)
	res := dgraph.SendMutation(question_info_mutation)
	if res.Success {
		res.Message = "Question Successfully Saved!"
	}
	question_json_response, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(question_json_response)
}

// FETCH All Questions HANDLER: Incomplete
func GetAllQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var ques GetQuestion
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	fmt.Println(ques)
	var question_mutation string
	if ques.Id != "" {
		question_mutation = "{debug(_xid_: rootQuestion) { question (after: " + ques.Id + ", first: 10) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	} else {
		question_mutation = "{debug(_xid_: rootQuestion) { question (first: 5) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	}
	fmt.Println(question_mutation)
	w.Header().Set("Content-Type", "application/json")

	question_response, err := http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(question_mutation))
	if err != nil {
		panic(err)
	}
	defer question_response.Body.Close()
	question_body, err := ioutil.ReadAll(question_response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(question_body))

	jsonResp, err := json.Marshal(string(question_body))
	if err != nil {
		panic(err)
	}

	w.Write(jsonResp)
}

// method to parse question response
func parseQuestionResponse(question_body []byte) (*QuestionAPIResponse, error) {
	var question_response = new(QuestionAPIResponse)
	err := json.Unmarshal(question_body, &question_response)
	if err != nil {
		fmt.Println("oops:", err)
	}
	return question_response, err
}

//API for "Adding Quiz" to Database
func AddQuizHandler(w http.ResponseWriter, r *http.Request) {
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
	stats := &Status{true, false, "Quiz Successfully Saved!"}
	quiz_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(quiz_json_response)
}

// fetch all the tags
func GetAllTagsHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	tag_mutation := "{debug(_xid_: rootQuestion) { question { question.tag { name _uid_} }}}"
	tag_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(tag_mutation))
	if err != nil {
		panic(err)
	}
	defer tag_response.Body.Close()
	tag_body, err := ioutil.ReadAll(tag_response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(tag_body))

	jsonResp, err := json.Marshal(string(tag_body))
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func EditQuestionHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var ques Question

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	fmt.Println(ques)
	// Creating query mutation to save question informations.
	question_info_mutation := "mutation { set { <_uid_:" + ques.Uid + "> <text> \"" + ques.Text +
		"\" . \n <_uid_:" + ques.Uid + "> <positive> \"" + strconv.FormatFloat(ques.Positive, 'g', -1, 64) +
		"\" . \n  <_uid_:" + ques.Uid + "> <negative> \"" + strconv.FormatFloat(ques.Negative, 'g', -1, 64) + "\" . \n "

	// Create and associate Options
	for l := range ques.Options {
		// index := strconv.Itoa(l)
		question_info_mutation += "<_uid_:" + ques.Options[l].Uid + "> <name> \"" + ques.Options[l].Name +
			"\" .\n <_uid_:" + ques.Uid + "> <question.option> <_uid_:" + ques.Options[l].Uid + "> . \n "

		// If this option is correct answer
		if ques.Options[l].Is_correct == true {
			fmt.Println(ques.Options[l])
			question_info_mutation += "<_uid_:" + ques.Uid + "> <question.correct> <_uid_:" + ques.Options[l].Uid + "> . \n "
		}
		if ques.Options[l].Is_correct == false {
			delete_correct := "mutation { delete { <_uid_:" + ques.Uid + "> <question.correct> <_uid_:" + ques.Options[l].Uid + "> .}}"
			_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(delete_correct))
			if err != nil {
				panic(err)
			}
		}
	}

	// Create and associate Tags
	for i := range ques.Tags {
		if ques.Tags[i].Uid != "" && ques.Tags[i].Is_delete == true {
			query_mutation := "mutation { delete { <_uid_:" + ques.Uid + "> <question.tag> <_uid_:" + ques.Tags[i].Uid + "> .}}"
			fmt.Println(query_mutation)
			_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(query_mutation))
			if err != nil {
				panic(err)
			}

		} else if ques.Tags[i].Uid != "" {
			question_info_mutation += "<_uid_:" + ques.Uid + "> <question.tag> <_uid_:" + ques.Tags[i].Uid +
				"> . \n <_uid_:" + ques.Tags[i].Uid + "> <tag.question> <_uid_:" + ques.Uid + "> . \n "
		} else {
			index := strconv.Itoa(i)
			question_info_mutation += "<_new_:tag" + index + "> <name> \"" + ques.Tags[i].Name +
				"\" .\n <_uid_:" + ques.Uid + "> <question.tag> <_new_:tag" + index +
				"> . \n <_new_:tag" + index + "> <tag.question> <_uid_:" + ques.Uid + "> . \n "
		}
	}

	question_info_mutation += " }}"
	fmt.Println(question_info_mutation)

	_, err = http.Post(dgraph.QueryEndpoint, "application/x-www-form-urlencoded", strings.NewReader(question_info_mutation))
	if err != nil {
		panic(err)
	}

	stats := &Status{true, false, "Question Successfully Updated!"}
	question_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(question_json_response)
}

func EditQuizHandler(w http.ResponseWriter, r *http.Request) {
	var quiz EditQuiz
	// var quizResp QuizAPIResponse

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
	stats := &Status{true, false, "Quiz Successfully Updated!"}
	quiz_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(quiz_json_response)
}

func GetAllQuizsHandler(w http.ResponseWriter, r *http.Request) {
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

// FILTER QUESTION HANDLER: Fileter By Tags
func FilterQuestionHandler(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	var tag TagFilter

	err := json.NewDecoder(r.Body).Decode(&tag)
	if err != nil {
		panic(err)
	}

	filter_query := "{root(_uid_: " + tag.UID + ") { tag.question { text }}"
	filter_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(filter_query))
	if err != nil {
		panic(err)
	}
	defer filter_response.Body.Close()
	filter_body, err := ioutil.ReadAll(filter_response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(filter_body))
	jsonResp, err := json.Marshal(string(filter_body))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func runHTTPServer(address string) {
	r := mux.NewRouter()
	// TODO - Change the API's to RESTful API's
	r.HandleFunc("/add-question", AddQuestionHandler).Methods("POST", "OPTIONS")
	// TODO - Change to PUT.
	r.HandleFunc("/edit-question", EditQuestionHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/get-all-questions", GetAllQuestionsHandler).Methods("POST", "OPTIONS")

	r.HandleFunc("/add-quiz", AddQuizHandler).Methods("POST", "OPTIONS")
	// TODO - Change to PUT.
	r.HandleFunc("/edit-quiz", EditQuizHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/get-all-quizes", GetAllQuizsHandler).Methods("GET", "OPTIONS")

	r.HandleFunc("/get-all-tags", GetAllTagsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/filter-questions", FilterQuestionHandler).Methods("POST", "OPTIONS")

	r.HandleFunc("/candidate", candidate.Add).Methods("POST", "OPTIONS")
	r.HandleFunc("/candidate/{id}", candidate.Edit).Methods("PUT", "OPTIONS")
	r.HandleFunc("/candidate/{id}", candidate.Get).Methods("GET", "OPTIONS")
	r.HandleFunc("/candidates", candidate.Index).Methods("GET", "OPTIONS")
	fmt.Println("Server Running on 8082")
	log.Fatal(http.ListenAndServe(address, r))
}

func main() {
	runHTTPServer(*port)
}
