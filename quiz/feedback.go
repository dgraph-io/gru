package quiz

import (
	"net/http"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
)

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
	}

	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + userId + `> <feedback> "` + f + `" .`)

	if _, err = dgraph.SendMutation(m.String()); err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
}
