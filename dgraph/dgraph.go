package dgraph

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	Server = flag.String("dgraph", "http://127.0.0.1:8080", "Dgraph server address")
	// TODO - Remove this later.
	QueryEndpoint = strings.Join([]string{*Server, "query"}, "/")
	endpoint      = strings.Join([]string{*Server, "query"}, "/")
)

type MutationRes struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Uids    map[string]string `json:"uids"`
}

type Mutation struct {
	set string
	del string
}

func (m *Mutation) Set(set string) {
	m.set += strings.Join([]string{m.set, set}, "\n")
}

func (m *Mutation) Del(del string) {
	m.del += strings.Join([]string{m.del, del}, "\n")
}

func (m *Mutation) String() string {
	var mutation string
	if len(m.set) > 0 {
		mutation += strings.Join([]string{"set {", m.set, "}"}, "\n")
	}
	if len(m.del) > 0 {
		mutation += strings.Join([]string{"\ndelete {", m.del, "}"}, "\n")
	}
	mutation = strings.Join([]string{"mutation {", mutation, "}"}, "\n")
	return mutation
}

func SendMutation(m string) MutationRes {
	res, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(m))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var mr MutationRes
	json.NewDecoder(res.Body).Decode(&mr)
	return mr
}

func Query(q string) []byte {
	res, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
