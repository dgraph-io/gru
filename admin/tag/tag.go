package tag

import (
	"net/http"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
)

type Tag struct {
	Uid       string `json:"_uid_"`
	Name      string `json:"name"`
	Is_delete bool
}

// fetch all the tags
// TODO - Clean this up.
func Index(w http.ResponseWriter, r *http.Request) {
	q := "{debug(_xid_: root) { question { question.tag { name _uid_} }}}"
	res, err := dgraph.Query(q)
	if err != nil {
		sr := server.Response{}
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
