package dgraph

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
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
	m.set = strings.Join([]string{m.set, set}, "\n")
}

func (m *Mutation) Del(del string) {
	m.del = strings.Join([]string{m.del, del}, "\n")
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

func SendMutation(m string) (MutationRes, error) {
	res, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(m))
	if err != nil {
		return MutationRes{}, errors.Wrap(err, "Couldn't send mutation")
	}
	defer res.Body.Close()

	var mr MutationRes
	json.NewDecoder(res.Body).Decode(&mr)
	if mr.Code != "ErrorOk" {
		return MutationRes{}, fmt.Errorf(mr.Message)
	}
	return mr, nil
}

func QueryAndUnmarshal(q string, i interface{}) error {
	res, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		return errors.Wrap(err, "Couldn't get response from Dgraph")
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Couldn't read response body")
	}
	if err = json.Unmarshal(b, i); err != nil {
		return err
	}
	return nil
}

func Query(q string) ([]byte, error) {
	res, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		return []byte{}, errors.Wrap(err, "Couldn't get response from Dgraph")
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Couldn't read response body")
	}
	return b, nil
}
