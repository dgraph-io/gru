package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	"github.com/dgraph-io/gru/x"
)

type pingRes struct {
	TimeLeft string `json:"time_left"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	var userId string
	var err error
	sr := server.Response{}
	if userId, err = validateToken(r); err != nil {
		sr.Write(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	if status, err := checkAndUpdate(userId); err != nil {
		sr.Write(w, "", err.Error(), status)
		return
	}

	c, err := readMap(userId)
	if err != nil {
		sr.Write(w, "", "Candidate not found.", http.StatusBadRequest)
		return
	}

	c.lastExchange = time.Now()
	updateMap(userId, c)
	pr := &pingRes{TimeLeft: "-1"}
	if !c.quizStart.IsZero() {
		end := c.quizStart.Add(c.quizDuration).Truncate(time.Second)
		timeLeft := end.Sub(time.Now().UTC().Truncate(time.Second))
		if timeLeft <= 0 {
			m := new(dgraph.Mutation)
			m.Set(`<_uid_:` + userId + `> <complete> "true" .`)
			m.Set(`<_uid_:` + userId + `> <score> "` + strconv.FormatFloat(x.ToFixed(c.score, 2), 'g', -1, 64) + `" .`)
			_, err := dgraph.SendMutation(m.String())
			if err != nil {
				sr.Write(w, "", err.Error(), http.StatusInternalServerError)
				return
			}
			if !c.mailSent {
				go sendReport(userId)
				c.mailSent = true
				updateMap(userId, c)
			}
		}
		pr.TimeLeft = timeLeft.String()
	}
	json.NewEncoder(w).Encode(pr)
}
