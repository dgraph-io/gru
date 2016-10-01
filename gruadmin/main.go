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
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dgraph-io/gru/gruadmin/candidate"
	"github.com/dgraph-io/gru/gruadmin/question"
	"github.com/dgraph-io/gru/gruadmin/quiz"
	"github.com/dgraph-io/gru/gruadmin/tag"
	"github.com/gorilla/mux"
)

var (
	port = flag.String("port", ":8082", "Port on which server listens")
)

func runHTTPServer(address string) {
	r := mux.NewRouter()
	// TODO - Change the API's to RESTful API's
	r.HandleFunc("/add-question", question.Add).Methods("POST", "OPTIONS")
	// TODO - Change to PUT.
	r.HandleFunc("/edit-question", question.Edit).Methods("POST", "OPTIONS")
	r.HandleFunc("/get-all-questions", question.Index).Methods("POST", "OPTIONS")
	r.HandleFunc("/filter-questions", question.Filter).Methods("POST", "OPTIONS")

	r.HandleFunc("/add-quiz", quiz.Add).Methods("POST", "OPTIONS")
	// TODO - Change to PUT.
	r.HandleFunc("/edit-quiz", quiz.Edit).Methods("POST", "OPTIONS")
	r.HandleFunc("/get-all-quizes", quiz.Index).Methods("GET", "OPTIONS")

	r.HandleFunc("/get-all-tags", tag.Index).Methods("GET", "OPTIONS")

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
