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
	Name string
}

type Question struct {
	Text     string
	Positive string
	Negative string
	Tags     []string
}

type Quiz struct {
	Name      string
	Duration  string
	StartDate string
	EndDate   string
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

// API for "Adding Question" to Database
func AddQuestionHandler(w http.ResponseWriter, r *http.Request) {
	var ques Question

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}
	//Running a Mutation - Generate question UID and attach to root node.
	question_mutation := "mutation { set { 	<root> <question> <_new_:question> . } }"
	question_response, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(question_mutation))
	if err != nil {
		panic(err)
	}

	question_body, err := ioutil.ReadAll(question_response.Body)
	if err != nil {
		panic(err.Error())
	}
	parsed_question_response, err := parseQuestionResponse([]byte(question_body)) //Parse Http respnse
	if err != nil {
		panic(err.Error())
	}
	strUID := strconv.FormatUint(parsed_question_response.UIDS.Question, 10)

	fmt.Println(strUID)
	for i := 0; i < len(ques.Tags); i++ {
		//fmt.Println(ques.Tags[i])
		tag_id := createTags(ques.Tags[i]) // call method to save the tag
		err = attachTagAndQuestion(strUID, tag_id)
		if err != nil {
			panic(err.Error())
		}
	}
	// Running mutation to save question informations.
	question_info_mutation := "mutation { set { 	<" + strUID + "> <text> \"" + ques.Text +
		"\" . \n <" + strUID + "> <positive> \"" + ques.Positive +
		"\" . \n <" + strUID + "> <negative> \"" + ques.Negative + "\" . " +
		" }}"
	question_info_response, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(question_info_mutation))
	if err != nil {
		panic(err)
	}
	question_info_body, err := ioutil.ReadAll(question_info_response.Body)
	if err != nil {
		panic(err.Error())
	}
	_, err = parseQuestionResponse([]byte(question_info_body)) //Parse Http respnse
	if err != nil {
		panic(err.Error())
	}
	question_json_response, err := json.Marshal(parsed_question_response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(question_json_response)
}

// Running mutation to save tag information
func createTags(tag string) (tag_id string ) {
	tag_mutation := "mutation { set {<_new_:tag> <name> \"" + tag + "\" . }}"
	tag_response, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(tag_mutation))
	if err != nil {
		panic(err)
	}

	tag_body, err := ioutil.ReadAll(tag_response.Body)
	if err != nil {
		panic(err.Error())
	}
	parsed_tag_response, err := parseTagResponse([]byte(tag_body)) //Parse Http respnse
	if err != nil {
		panic(err)
	}
	strTagUID := strconv.FormatUint(parsed_tag_response.UIDS.Tag, 10)
	fmt.Println(strTagUID)
	return strTagUID
}

// Running mutation to establish relationship between question and tag
func attachTagAndQuestion(question_id string, tag_id string) (error){
	relationship_mutation := "mutation { set { 	<" + question_id + "> <question.tag> <" + tag_id +
		"> . \n <" + tag_id + "> <tag.question> <" + question_id +
		"> . }}"
	relationship_response, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(relationship_mutation))
	if err != nil {
		panic(err)
	}
	relationship_body, err := ioutil.ReadAll(relationship_response.Body)
	if err != nil {
		panic(err.Error())
	}
	_, err = parseTagResponse([]byte(relationship_body)) //Parse Http respnse
	if err != nil {
		panic(err.Error())
	}
	return err
}

// method to parse tag response
func parseTagResponse(tag_body []byte) (*TagAPIResponse, error) {
	var tag_response = new(TagAPIResponse)
	err := json.Unmarshal(tag_body, &tag_response)
	if err != nil {
		fmt.Println("oops:", err)
	}
	return tag_response, err
}

// FETCH All Questions HANDLER: Incomplete
func GetAllQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	question_mutation := "{root(_xid_: root) { question { text negative positive  }  } }"
	question_response, _ := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(question_mutation))
	defer question_response.Body.Close()
	question_body, err := ioutil.ReadAll(question_response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(question_body))

	// data := map[string]interface{}{}
	// dec := json.NewDecoder(body)
	// dec.Decode(&data)

	// jsonResp, err := json.Marshal(data)

	// fmt.Println("start")
	// fmt.Println(jsonResp)
	// fmt.Println("end")

	w.Header().Set("Content-Type", "application/json")
	// w.Write(jsonResp)
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

//API for "Adding Tags" to Database
func AddTagsHandler(w http.ResponseWriter, r *http.Request) {
	var tag Tag
	err := json.NewDecoder(r.Body).Decode(&tag)
	if err != nil {
		panic(err)
	}
	_ = createTags(tag.Name) // method to save the tag
	stats := &Status{true, false, "Tag Successfully Saved!"}
	tag_json_response, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(tag_json_response)
}

//API for "Adding Quiz" to Database
func AddQuizHandler(w http.ResponseWriter, r *http.Request) {
	var quiz Quiz
	var quizResp QuizAPIResponse

	err := json.NewDecoder(r.Body).Decode(&quiz)
	if err != nil {
		panic(err)
	}
	quiz_mutation := "mutation { set { 	<_new_:quiz> <name> \"" + quiz.Name +
		"\" . \n <_new_:quiz> <duration> \"" + quiz.Duration +
		"\" . \n <_new_:quiz> <start-date> \"" + quiz.StartDate + "\" . " +
		"\" . \n <_new_:quiz> <end-date> \"" + quiz.EndDate + "\" . " +
		" }}"
	quiz_response, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_mutation))
	if err != nil {
		panic(err)
	}

	quiz_body, err := ioutil.ReadAll(quiz_response.Body)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(quiz_body, &quizResp)
	if err != nil {
		panic(err)
	}

	// Query to fetch stored Quix, display on console for Cross check
	quiz_query := "{root(_uid_:" + strconv.FormatUint(quizResp.UIDS.Quiz, 10) + ") { name duration start-date _uid_ }}"
	query_response, _ := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(quiz_query))
	defer query_response.Body.Close()
	query_body, err := ioutil.ReadAll(query_response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(query_body))

	quiz_json_response, err := json.Marshal(quizResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(quiz_json_response)
}

// FETCH TAGS HANDLER: Incomplete
func GetTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Need to change the hardcoded tag UID
	q := "{root(_xid_:root) { question { text positive } }}"
	// q := "{root(_xid_:" + strconv.FormatUint(3773063241899813845, 10) + ") { text positive question.tag { name _uid_ } _uid_ }}"
	//q := "{question(_uid_:" + strconv.FormatUint(2878207119269946827, 10) + ") {question.tag {name _uid_ }}"
	resp, _ := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(q))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	jsonResp, err := json.Marshal(string(body))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func runHTTPServer(address string) {
	http.HandleFunc("/add-tag", AddTagsHandler)
	http.HandleFunc("/add-question", AddQuestionHandler)
	http.HandleFunc("/add-quiz", AddQuizHandler)
	http.HandleFunc("/get-tag", GetTagsHandler)
	http.HandleFunc("/get-all-questions", GetAllQuestionsHandler)

	fmt.Println("Server Running on 8082")
	log.Fatal(http.ListenAndServe(address, nil))
}

func main() {
	runHTTPServer(*port)
}