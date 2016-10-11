package tag

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgraph-io/gru/gruadmin/server"
	"github.com/dgraph-io/gru/x"
)

type Tag struct {
	Uid       string `json:"_uid_"`
	Name      string `json:"name"`
	Is_delete bool
}

// fetch all the tags
func Index(w http.ResponseWriter, r *http.Request) {
	server.AddCorsHeaders(w)
	tag_mutation := "{debug(_xid_: rootQuestion) { question { question.tag { name _uid_} }}}"
	tag_response, err := http.Post("http://localhost:8080/query", "application/x-www-form-urlencoded", strings.NewReader(tag_mutation))
	if err != nil {
		panic(err)
	}
	defer tag_response.Body.Close()
	tag_body, err := ioutil.ReadAll(tag_response.Body)
	if err != nil {
		panic(err)
	}
	x.Debug(string(tag_body))

	jsonResp, err := json.Marshal(string(tag_body))
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
