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
	// "reflect"

	//"golang.org/x/net/context"
	"google.golang.org/grpc"

)

type Tag struct {
	Id string
	Name string
}

type Option struct {
		Name string
    Is_correct bool
	}

type Question struct {
	Text     string
	Positive float64 
	Negative float64
	Tags     []Tag
	Options []Option
}

type EditQuestion struct {
	Id string
	Text     string
	Positive float64 
	Negative float64
	Tags     []Tag
	Options []Option
}

type Quiz struct {
	Name      string
	Duration  string
	Start_Date string
	End_Date   string
	Questions []string
}

type EditQuiz struct {
	Id string
	Name      string
	Duration  string
	Start_Date string
	End_Date   string
	Questions []string
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

var (
	port      = flag.String("port", ":8082", "Port on which server listens")
	graphPort = flag.String("graphPort", ":8080", "Port on which dgraph listens")
	conn      *grpc.ClientConn
	endPoint  = "http://localhost:8080/query"
)

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers",
		"Authorization,Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token,"+
			"X-Auth-Token, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Connection", "close")
}

// API for "Adding Question" to Database
func AddQuestionHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w)
	var ques Question

	if r.Method == "OPTIONS" {
		return
	}

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}

	// // Creating query mutation to save question informations.
	question_info_mutation := "mutation { set { <rootQuestion> <question> <_new_:question> . \n	<_new_:question> <text> \"" + ques.Text +
		"\" . \n <_new_:question> <positive> \"" + strconv.FormatFloat(ques.Positive, 'g', -1, 64) +
		"\" . \n <_new_:question> <negative> \"" + strconv.FormatFloat(ques.Negative, 'g', -1, 64) + "\" . \n " 

		// Create and associate Options
		for l := range ques.Options {
			fmt.Println(ques.Options[l].Is_correct)
			index := strconv.Itoa(l)
			question_info_mutation += "<_new_:option"+index+ "> <name> \"" + ques.Options[l].Name + 
			"\" .\n <_new_:question> <question.option> <_new_:option"+index+ "> . \n "

			// If this option is correct answer
			if ques.Options[l].Is_correct == true {
				question_info_mutation += "<_new_:question> <question.correct> <_new_:option"+index+ "> . \n "
			}		
    }

    // Create and associate Tags
    for i := range ques.Tags {
    	if ques.Tags[i].Id != "" {
    		question_info_mutation += "<_new_:question> <question.tag> <_uid_:"+ ques.Tags[i].Id + 
				"> . \n <_uid_:"+ ques.Tags[i].Id + "> <tag.question> <_new_:question> . \n "
    	} else {
    		index := strconv.Itoa(i)
				question_info_mutation += "<_new_:tag"+index+ "> <name> \"" + ques.Tags[i].Name + 
				"\" .\n <_new_:question> <question.tag> <_new_:tag"+ index + 
				"> . \n <_new_:tag"+ index + "> <tag.question> <_new_:question> . \n "
    	}
		}

	question_info_mutation +=	" }}"
	fmt.Println(question_info_mutation)
	
	_, err = http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(question_info_mutation))
	if err != nil {
		panic(err)
	}
	
	stats := &Status{true, false, "Question Successfully Saved!"}
	question_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(question_json_response)
}


// FETCH All Questions HANDLER: Incomplete
func GetAllQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	addCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var ques GetQuestion

	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	fmt.Println(ques)
	addCorsHeaders(w)
	var question_mutation string
	if ques.Id != "" {
		question_mutation = "{debug(_xid_: rootQuestion) { question (after: " + ques.Id + ", first: 10) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	} else {
		question_mutation = "{debug(_xid_: rootQuestion) { question (first: 10) { _uid_ text negative positive question.tag { name } question.option { name } question.correct { name } }  } }"
	}	
	fmt.Println(question_mutation)
	question_response, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(question_mutation))
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

	w.Header().Set("Content-Type", "application/json")
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
	addCorsHeaders(w)
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
	_, err = http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
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
	addCorsHeaders(w)
	tag_mutation := "{debug(_xid_: rootQuestion) { question { question.tag { name _uid_} }}}"
	tag_response, _ := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(tag_mutation))
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
	var ques EditQuestion

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	fmt.Println(ques)
	// Creating query mutation to save question informations.
	question_info_mutation := "mutation { set { <_uid_:" + ques.Id + "> <text> \"" + ques.Text +
		"\" . \n <_uid_:" + ques.Id + "> <positive> \"" + strconv.FormatFloat(ques.Positive, 'g', -1, 64) +
		"\" . \n  <_uid_:" + ques.Id + "> <negative> \"" + strconv.FormatFloat(ques.Negative, 'g', -1, 64) + "\" . \n " 

	
		// Create and associate Options
		for l := range ques.Options {
			index := strconv.Itoa(l)
			question_info_mutation += "<_new_:option"+index+ "> <name> \"" + ques.Options[l].Name + 
			"\" .\n <_uid_:" + ques.Id + "> <question.option> <_new_:option"+index+ "> . \n "

			// If this option is correct answer
			if ques.Options[l].Is_correct == true {
				question_info_mutation += "<_uid_:" + ques.Id + "> <question.correct> <_new_:option"+index+ "> . \n "
			}		
    }

    // Create and associate Tags
    for i := range ques.Tags {
    	if ques.Tags[i].Id != "" {
    		question_info_mutation += "<_uid_:" + ques.Id + "> <question.tag> <_uid_:"+ ques.Tags[i].Id + 
				"> . \n <_uid_:"+ ques.Tags[i].Id + "> <tag.question> <_uid_:" + ques.Id + "> . \n "
    	} else {
    		index := strconv.Itoa(i)
				question_info_mutation += "<_new_:tag"+index+ "> <name> \"" + ques.Tags[i].Name + 
				"\" .\n <_uid_:" + ques.Id + "> <question.tag> <_new_:tag"+ index + 
				"> . \n <_new_:tag"+ index + "> <tag.question> <_uid_:" + ques.Id + "> . \n "
    	}
		}

	question_info_mutation +=	" }}"
	fmt.Println(question_info_mutation)
	
	_, err = http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(question_info_mutation))
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
	_, err = http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
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
	addCorsHeaders(w)
	quiz_mutation := "{debug(_xid_: rootQuiz) { quiz { _uid_ name duration quiz.question { text }  }  }}"
	quiz_response, _ := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
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

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func runHTTPServer(address string) {
	//http.HandleFunc("/add-tag", AddTagsHandler)
	http.HandleFunc("/add-question", AddQuestionHandler)
	http.HandleFunc("/add-quiz", AddQuizHandler)
	http.HandleFunc("/get-all-tags", GetAllTagsHandler)
	http.HandleFunc("/get-all-questions", GetAllQuestionsHandler)
	http.HandleFunc("/edit-question", EditQuestionHandler)
	http.HandleFunc("/edit-quiz", EditQuizHandler)
	http.HandleFunc("/get-all-quizes", GetAllQuizsHandler)

	fmt.Println("Server Running on 8082")
	log.Fatal(http.ListenAndServe(address, nil))
}

func main() {
	runHTTPServer(*port)
}