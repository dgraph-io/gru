package quiz

import (
	"net/http"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
)

// TODO - Later remove this when we have a proxy endpoint for quiz candidates too.
func Feedback(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	userId, err := validateToken(r)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	f := r.PostFormValue("feedback")
	if f == "" {
		sr.Write(w, "", "Feedback can't be empty.", http.StatusBadRequest)
		return
	}

	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + userId + `> <feedback> "` + f + `" .`)

	if _, err = dgraph.SendMutation(m.String()); err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
}
