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

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dgraph-io/dgraphgoclient/graph"
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

	//Running a Mutation - Saving Question Infos, Get UID
	q := "mutation { set { 	<rootQuestion> <question> <_new_:question> . } }"
	resp, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	s, err := parseResponse([]byte(body)) //Parse Http respnse
	strUID := strconv.FormatUint(s.UIDS.Question, 10)

	fmt.Println(strUID)
	// create tag

	for i := 0; i < len(ques.Tags); i++ {
		fmt.Println(ques.Tags[i])
		tags, err := createTags(ques.Tags[i], strUID)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(tags)
	}

	q = "mutation { set { 	<" + strUID + "> <text> \"" + ques.Text +
		"\" . \n <" + strUID + "> <positive> \"" + ques.Positive +
		"\" . \n <" + strUID + "> <negative> \"" + ques.Negative + "\" . " +
		" }}"
	resp1, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		panic(err)
	}
	body2, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		panic(err.Error())
	}
	s1, err := parseResponse([]byte(body2)) //Parse Http respnse
	fmt.Println(s1)

	jsonResp, err := json.Marshal(s)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func createTags(tag string, strUID string) (*TagAPIResponse, error) {
	q := "mutation { set {<_new_:tag> <name> \"" + tag + "\" . }}"
	resp, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	s, err := parseTagResponse([]byte(body)) //Parse Http respnse
	strTagUID := strconv.FormatUint(s.UIDS.Tag, 10)
	fmt.Println(strTagUID)

	q = "mutation { set { 	<" + strUID + "> <question.tag> \"" + strTagUID +
		"\" . \n <" + strTagUID + "> <tag.question> \"" + strUID +
		"\" . }}"
	resp1, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		panic(err)
	}
	body2, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		panic(err.Error())
	}
	s1, err := parseTagResponse([]byte(body2)) //Parse Http respnse
	return s1, err
}

func parseTagResponse(body []byte) (*TagAPIResponse, error) {
	var s = new(TagAPIResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("oops:", err)
	}
	return s, err
}

// FETCH All Questions HANDLER: Incomplete
func GetAllQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	q := "{root(_xid_:rootQuestion) { question { text negative positive _uid_ }  } }"
	httpRes, _ := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	defer httpRes.Body.Close()
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

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

func parseResponse(body []byte) (*QuestionAPIResponse, error) {
	var s = new(QuestionAPIResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("oops:", err)
	}
	return s, err
}

//API for "Adding Tags" to Database
func AddTagsHandler(w http.ResponseWriter, r *http.Request) {
	var t Tag

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		panic(err)
	}

	c := graph.NewDgraphClient(conn)

	//Running a Mutation
	q := "mutation { set {<_new_:tag> <name> \"" + t.Name + "\" . }}"
	resp, err := c.Query(context.Background(), &graph.Request{Query: q})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	stats := &Status{true, false, "Tag Successfully Saved!"}
	jsonResp, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

//API for "Adding Quiz" to Database
func AddQuizHandler(w http.ResponseWriter, r *http.Request) {
	var quiz Quiz
	var quizResp QuizAPIResponse

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&quiz)
	if err != nil {
		panic(err)
	}

	//Running a Mutation - Saving Question Infos
	q := "mutation { set { 	<_new_:quiz> <name> \"" + quiz.Name +
		"\" . \n <_new_:quiz> <duration> \"" + quiz.Duration +
		"\" . \n <_new_:quiz> <start-date> \"" + quiz.StartDate + "\" . " +
		"\" . \n <_new_:quiz> <end-date> \"" + quiz.EndDate + "\" . " +
		" }}"
	resp, err := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(body, &quizResp)
	if err != nil {
		panic(err)
	}

	//Dgraph Query :  Returning Stored Quiz as response for Cross Check
	q = "{root(_uid_:" + strconv.FormatUint(quizResp.UIDS.Quiz, 10) + ") { name duration start-date _uid_ }}"
	httpRes, _ := http.Post(endPoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	defer httpRes.Body.Close()
	body1, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body1))

	jsonResp, err := json.Marshal(quizResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

// FETCH TAGS HANDLER: Incomplete
func GetTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Need to change the hardcoded tag UID
	q := "{root(_uid_:" + strconv.FormatUint(11971855182167083540, 10) + ") { name _uid_ }}"
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

// INITIALIZERS
func connectDgraphClient() {
	var err error
	conn, err = grpc.Dial(*graphPort, grpc.WithInsecure())
	if err != nil {
		log.Fatal("DialTCPConnection")
	}
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

func init() {
	connectDgraphClient()
}

func main() {
	runHTTPServer(*port)
}
