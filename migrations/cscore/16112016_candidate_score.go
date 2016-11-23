// This is used to populate the total score fields for each candidate.
// From now on it would automatically be stored when the quiz for a candidate ends.
package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
)

var quiz = flag.String("quiz", "", "Quiz id")

type Quiz struct {
	Quizzes []struct {
		Cand []struct {
			Id       string `json:"_uid_"`
			Complete bool   `json:",string"`
			Question []struct {
				Score float64 `json:"candidate.score,string"`
			} `json:"candidate.question"`
			Score float64
		} `json:"quiz.candidate"`
	} `json:"quiz"`
}

func candidates() string {
	return `{
        quiz(_uid_: ` + *quiz + `) {
                quiz.candidate {
                        _uid_
                        complete
                        candidate.question {
                                candidate.score
                        }
                }
        }
}`
}

func main() {
	flag.Parse()
	if *quiz == "" {
		log.Fatal("Quiz can't be empty")
	}
	q := candidates()
	var qu Quiz
	dgraph.QueryAndUnmarshal(q, &qu)
	if len(qu.Quizzes) == 0 || len(qu.Quizzes[0].Cand) == 0 {
		log.Fatal("Couldn't find candidate data for the quiz.")
	}
	cand := qu.Quizzes[0].Cand
	for _, c := range cand {
		if !c.Complete {
			continue
		}
		score := 0.0
		for _, qn := range c.Question {
			score += qn.Score
		}
		c.Score = x.ToFixed(score, 2)
		m := new(dgraph.Mutation)
		m.Set(`<_uid_:` + c.Id + `> <score> "` + strconv.FormatFloat(c.Score, 'g', -1, 64) + `" .`)
		if _, err := dgraph.SendMutation(m.String()); err != nil {
			log.Fatalf("Error: %v for candidate with uid: %v", err, c.Id)
		}
	}
}
