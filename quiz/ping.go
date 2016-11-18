package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
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
	// If quiz hasn't started yet, we return time_left as -1.
	if c.quizStart.IsZero() {
		json.NewEncoder(w).Encode(pr)
		return
	}

	end := c.quizStart.Add(c.quizDuration).Truncate(time.Second)
	timeLeft := end.Sub(time.Now().UTC().Truncate(time.Second))
	pr.TimeLeft = timeLeft.String()
	if timeLeft > 0 {
		json.NewEncoder(w).Encode(pr)
		return
	}

	// Time left is <=0, that means quiz should end now. Lets store this information.
	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + userId + `> <complete> "true" .`)
	m.Set(`<_uid_:` + userId + `> <score> "` + strconv.FormatFloat(x.ToFixed(c.score, 2), 'g', -1, 64) + `" .`)
	_, err = dgraph.SendMutation(m.String())
	if err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	// Client may call ping twice after the timeLeft <= 0, but we wan't to send mail
	// only once. So we check if we already sent the mail.
	if !c.mailSent {
		if c.score <= c.quizCutoff {
			go mail.Reject(c.name, c.email)
		}
		go sendReport(userId)
		c.mailSent = true
		updateMap(userId, c)
	}
	json.NewEncoder(w).Encode(pr)
}
