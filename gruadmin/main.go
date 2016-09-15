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
}

type Status struct { // HTTP Response status
	Success bool
	Error   bool
	Message string
}

type DgraphResponse struct { // HTTP Response status
	Code    string
	Message string
	UIDS    struct {
		Question string
	}
}

var (
	port      = flag.String("port", ":8082", "Port on which server listens")
	graphPort = flag.String("graphPort", ":8080", "Port on which dgraph listens")
	conn      *grpc.ClientConn
)

func AddQuestionHandler(w http.ResponseWriter, r *http.Request) {
	var ques Question
	var dresp DgraphResponse

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&ques)
	if err != nil {
		panic(err)
	}

	//Running a Mutation - Saving Question Infos
	q := "mutation { set { <_new_:question> <text> \"" + ques.Text +
		"\" . \n <_new_:question> <positive> \"" + ques.Positive +
		"\" . \n <_new_:question> <negative> \"" + ques.Negative + "\" . " +
		" }}"
	resp, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(dresp)
	fmt.Println(dresp.Code)
	fmt.Println(dresp.UIDS.Question)

	//Dgraph Query :  Returning Stored Question as response for Cross Check
	q = "{root(_uid_:" + dresp.Code + ") { text negative positive _uid_ }}"
	httpRes, _ := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(q))
	defer httpRes.Body.Close()
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	jsonResp, err := json.Marshal(string(body))

	// stats := &Status{true, false, "Question Successfully Saved!"}
	// jsonResp, err := json.Marshal(jsonResp)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

// ADD NEW TAGS HANDLER
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

// FETCH TAGS HANDLER: Incomplete
func GetTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Need to change the hardcoded tag UID
	q := "{root(_uid_:" + strconv.FormatUint(11971855182167083540, 10) + ") { name _uid_ }}"

	resp, _ := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(q))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	jsonResp, err := json.Marshal(string(body))

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
	http.HandleFunc("/get-tag", GetTagsHandler)

	fmt.Println("Server Running on 8082")
	log.Fatal(http.ListenAndServe(address, nil))
}

func init() {
	connectDgraphClient()
}

func main() {
	runHTTPServer(*port)
}
