// This migration changes the quiz duration stored to number of minutes instead of
// a go time.Duration string.
// Earlier duration was stored as "1h40m0s", we want to change it to just "100".
package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/gru/dgraph"
)

type res struct {
	Root []struct {
		Quizzes []struct {
			Id       string `json:"_uid_"`
			Duration string `json:"duration"`
		} `json:"quiz"`
	} ` json:"root"`
}

func quizzes() string {
	return `{
        root(id: root) {
                quiz {
                        _uid_
                        duration
                }
        }
}`
}

func main() {
	q := quizzes()
	var res res
	dgraph.QueryAndUnmarshal(q, &res)
	if len(res.Root) == 0 || len(res.Root[0].Quizzes) == 0 {
		log.Fatal("No quizzes found.")
	}
	for _, quiz := range res.Root[0].Quizzes {
		t, err := time.ParseDuration(quiz.Duration)
		if err != nil {
			fmt.Printf("Couldn't convert duration: %v. Got err: %v",
				quiz.Duration, err)
		}
		m := new(dgraph.Mutation)
		m.Set(`<` + quiz.Id + `> <duration> "` + strconv.FormatFloat(t.Minutes(), 'g', -1, 64) + `" .`)
		if _, err := dgraph.SendMutation(m.String()); err != nil {
			log.Fatalf("Error: %v while performing mutation for quiz with uid: %v", err, quiz.Id)
		}
	}
}
