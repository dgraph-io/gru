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

package quiz

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	quizmeta "github.com/dgraph-io/gru/gruserver/quiz"
)

type Candidate struct {
	name     string
	email    string
	validity time.Time
	score    float32
	// List of quiz questions that have not been asked to the candidate yet.
	questions []Question
	// count of demo qns asked.
	demoQnsAsked int
	demoQnsToAsk int
	demoTaken    bool
	logFile      *os.File
	demoStart    time.Time
	quizStart    time.Time
	// session id of currently active session.
	sid          string
	lastExchange time.Time
}

var (
	tls      = flag.Bool("tls", true, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "fullchain.pem", "The TLS cert file")
	keyFile  = flag.String("key_file", "privkey.pem", "The TLS key file")
	quizFile = flag.String("quiz", "quiz.yml", "Input question file")
	port     = flag.String("port", ":443", "Port on which server listens")
	candFile = flag.String("cand", "candidates.txt", "Candidate inforamation")
	// This is the number of demo questions asked to dummy candidates.
	maxDemoQns = flag.Int("max_demo_qns", 25, "Maximum number of demo questions for dummy candidates.")
	// List of question ids.
	questions []Question
	cmap      map[string]Candidate
	wrtLock   sync.Mutex
	mapLock   sync.Mutex
	throttle  = make(chan time.Time, 3)
	rate      = time.Second
)

type server struct{}

func checkToken(c Candidate) error {
	if time.Now().UTC().After(c.validity) {
		return errors.New("Your token has expired.")
	}
	// Initially quizStart is zero, but after candidate has taken the
	// quiz once, it shouldn't be zero.
	if !c.quizStart.IsZero() && time.Now().UTC().After(c.quizStart.Add(quizDuration)) {
		// TODO - Show duration elapsed in minutes.
		return fmt.Errorf(
			"%v since you started the quiz for the first time are already over",
			quizDuration)
	}
	return nil
}

func updateMap(token string, c Candidate) {
	mapLock.Lock()
	defer mapLock.Unlock()
	cmap[token] = c
}

func readMap(token string) (Candidate, bool) {
	mapLock.Lock()
	defer mapLock.Unlock()
	c, ok := cmap[token]
	return c, ok
}

func sliceDiff(qnList []Question, qnsAsked []string) []Question {
	qns := []Question{}
	qmap := make(map[string]bool)
	for _, q := range qnsAsked {
		qmap[q] = true
	}
	for _, q := range qnList {
		if present := qmap[q.Id]; !present {
			qns = append(qns, q)
		}
	}
	return qns
}

func shuffleOptions(opts []*quizmeta.Answer) {
	for i := range opts {
		j := rand.Intn(i + 1)
		opts[i], opts[j] = opts[j], opts[i]
	}
}

func shuffleQuestions(qns []Question) {
	for i := range qns {
		j := rand.Intn(i + 1)
		qns[i], qns[j] = qns[j], qns[i]
	}
}

func authenticate(token string) (s Session, err error) {
	var c Candidate
	var ok bool
	if strings.HasPrefix(token, "test-") {
		c = demoCandInfo(token)
		s = Session{
			Id:           RandStringBytes(36),
			State:        QUIZ_DEMO_NOT_TAKEN,
			TestDuration: quizDuration.String(),
			DemoDuration: demoDuration.String(),
		}
		if !c.demoStart.IsZero() {
			s.TimeLeft = timeLeft(demoDuration, c.demoStart).String()
		} else {
			s.TimeLeft = demoDuration.String()
		}
		writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), s.Id))
		return
	}
	if c, ok = readMap(token); !ok {
		err = errors.New("Invalid token")
		return
	}

	if c, err = candInfo(token, c); err != nil {
		return
	}
	if err = checkToken(c); err != nil {
		return
	}

	timeSinceLastExchange := time.Now().Sub(c.lastExchange)
	if !c.lastExchange.IsZero() && timeSinceLastExchange < 10*time.Second {
		fmt.Println("Duplicate session for same auth token", token, c.name)
		err = errors.New("Duplicate Session. You already have an open session. If not try after 10 seconds.")
		return
	}

	s = Session{
		Id:           RandStringBytes(36),
		State:        state(c),
		TestDuration: quizDuration.String(),
		DemoDuration: demoDuration.String(),
	}

	if state(c) == QUIZ_DEMO_NOT_TAKEN || state(c) == QUIZ_DEMO_STARTED {
		s.TimeLeft = timeLeft(demoDuration, c.demoStart).String()
	} else {
		s.TimeLeft = timeLeft(quizDuration, c.quizStart).String()
	}
	writeLog(c, fmt.Sprintf("%v session_token %s\n", UTCTime(), s.Id))
	c.sid = s.Id
	c.lastExchange = time.Now()
	updateMap(token, c)
	return
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	var s Session
	var err error

	select {
	case <-throttle:
		s, err = authenticate(r.Header.Get("Authorization"))
	case <-time.After(time.Second * 1):
		err = errors.New("Please try again later. Too much load on server.")
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s)
}

func timeLeft(dur time.Duration, ts time.Time) time.Duration {
	return (dur - time.Now().UTC().Sub(ts))
}

// func main() {
// 	rand.Seed(time.Now().UTC().UnixNano())
// 	flag.Parse()
// 	cmap = make(map[string]Candidate)
// 	go rateLimit()
// 	runHTTPServer(*port)
// }
