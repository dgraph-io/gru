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
	"log"
	"net/http"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dgraph-io/dgraphgoclient/graph"
)

type tag struct {
	Name string
}

type status struct {
	message string
	error string
}

var (
	port      = flag.String("port", ":8082", "Port on which server listens")
	graphPort = flag.String("graphPort", ":8080", "Port on which dgraph listens")
	conn      *grpc.ClientConn
)

func TagsHandler(w http.ResponseWriter, r *http.Request) {
	var t tag

	// Decoding post data
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		panic(err)
	}

	//Running a Mutation
	c := graph.NewDgraphClient(conn)
	q := "mutation { set {<_new_:tag> <name> \"" + t.Name + "\" . }}"
	resp, err := c.Query(context.Background(), &graph.Request{Query: q})
	if err != nil {
		panic(err)
	}

	// Fetching the data
	q = "{root(_uid_:"+ strconv.FormatUint(resp.AssignedUids["tag"], 10) +") { name _uid_ }}"
	resp, err = c.Query(context.Background(), &graph.Request{Query: q})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.N)
  jsonResp, err := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func connectDgraphClient() {
	var err error
	conn, err = grpc.Dial(*graphPort, grpc.WithInsecure())
	if err != nil {
		log.Fatal("DialTCPConnection")
	}
}

func runHTTPServer(address string) {
	http.HandleFunc("/add-tag", TagsHandler)
	// http.HandleFunc("/get-tag", GetTags)

	fmt.Println("Server Running on 8082")
	log.Fatal(http.ListenAndServe(address, nil))
}

func init() {
	connectDgraphClient()
}

func main() {
	runHTTPServer(*port)
}

