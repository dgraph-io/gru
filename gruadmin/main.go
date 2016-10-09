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

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgraph-io/gru/auth"
	"github.com/dgraph-io/gru/gruadmin/candidate"
	"github.com/dgraph-io/gru/gruadmin/question"
	"github.com/dgraph-io/gru/gruadmin/quiz"
	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/gruadmin/tag"
	"github.com/dgraph-io/gru/x"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

var (
	port     = flag.String("port", ":8082", "Port on which server listens")
	username = flag.String("user", "", "Username to login to admin panel")
	password = flag.String("pass", "", "Username to login to admin panel")
)

func login(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}
	u, p, ok := r.BasicAuth()
	if !ok || u != *username || p != *password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// TODO - Add relevant claims like expiry.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})

	tokenString, err := token.SignedString([]byte(*auth.Secret))
	if err != nil {
		log.Fatal(err)
	}
	x.Debug(tokenString)
	w.Header().Set("Content-Type", "application/json")

	type Res struct {
		Token string `json:"token"`
	}

	res := Res{Token: tokenString}
	json.NewEncoder(w).Encode(res)
}

func runHTTPServer(address string) {
	router := mux.NewRouter()

	router.HandleFunc("/login", login).Methods("POST", "OPTIONS")
	router.HandleFunc("/quiz/{id}", candidate.Validate).Methods("POST", "OPTIONS")

	adminRouter := mux.NewRouter().PathPrefix("/admin").Subrouter().StrictSlash(true)

	// TODO - Change the API's to RESTful API's
	adminRouter.HandleFunc("/add-question", question.Add).Methods("POST", "OPTIONS")
	// TODO - Change to PUT.
	adminRouter.HandleFunc("/get-all-questions", question.Index).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/filter-questions", question.Filter).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/question/{id}", question.Get).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/question/{id}", question.Edit).Methods("PUT", "OPTIONS")

	adminRouter.HandleFunc("/add-quiz", quiz.Add).Methods("POST", "OPTIONS")
	// TODO - Change to PUT.
	adminRouter.HandleFunc("/get-all-quizes", quiz.Index).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/quiz/{id}", quiz.Get).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/quiz/{id}", quiz.Edit).Methods("PUT", "OPTIONS")

	adminRouter.HandleFunc("/get-all-tags", tag.Index).Methods("GET", "OPTIONS")

	adminRouter.HandleFunc("/candidate", candidate.Add).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/candidate/{id}", candidate.Edit).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/candidate/{id}", candidate.Get).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/candidates", candidate.Index).Methods("GET", "OPTIONS")

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(*auth.Secret), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	router.PathPrefix("/admin").Handler(negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(adminRouter),
	))
	n := negroni.Classic()
	n.UseHandler(router)
	fmt.Println("Server Running on 8082")
	log.Fatal(http.ListenAndServe(address, n))
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runHTTPServer(*port)
}
