package dgraph

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
  "google.golang.org/grpc"
	"github.com/dgraph-io/gru/admin/server"
	"github.com/pkg/errors"
)

var (
	Server = flag.String("dgraph", "127.0.0.1:9080", "Dgraph server address")
	// TODO switch to 100% dgo & grpc
	HttpServer = flag.String("httpdgraph", "http://127.0.0.1:8080", "Dgraph HTTP address")
)

const Success = "Success"

func _newClient() *dgo.Dgraph {
	d, err := grpc.Dial(*Server, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}

var _dgClient *dgo.Dgraph

func getDgraphClient() *dgo.Dgraph {
	if _dgClient == nil {
		_dgClient = _newClient()
	}
	return _dgClient
}

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

func (m *Mutation) SetString(l string, p string, val string) {
	m.Set(`<` + l + `> <` + p + `> "` + val + `" .`)
}

func (m *Mutation) SetLink(l string, p string, val string) {
	m.Set(`<` + l + `> <` + p + `> <` + val + `> .`)
}

func (m *Mutation) DelLink(l string, p string, val string) {
	m.Del(`<` + l + `> <` + p + `> <` + val + `> .`)
}

func (m *Mutation) Del(del string) {
	m.del = strings.Join([]string{m.del, del}, "\n")
}

func SendMutation(m *Mutation) (MutationRes, error) {
	txn := getDgraphClient().NewTxn()
	ctx := context.Background()
	defer txn.Discard(ctx)

	res, err := txn.Mutate(ctx, &api.Mutation{
		SetNquads: []byte(m.set), DelNquads: []byte(m.del) })
	if err != nil {
		return MutationRes{}, err
	}

	var mr MutationRes
	mr.Uids = res.Uids
	mr.Code = Success

	err = txn.Commit(ctx)
	if err != nil {
		return MutationRes{}, err
	}
	return mr, nil
}

func QueryAndUnmarshal(q string, i interface{}) error {
	endpoint := strings.Join([]string{*HttpServer, "query"}, "/")
	res, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(q))
	if err != nil {
		return errors.Wrap(err, "Couldn't get response from Dgraph")
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "Couldn't read response body")
	}
	fmt.Println("Query:", q)
	fmt.Println("Body: ", string(b))
	if err = json.Unmarshal(b, i); err != nil {
		fmt.Println("Failed to Unmarshal ", err)
		return err
	}
	fmt.Println("Parsed as ", i)
	return nil
}

func Query(q string) ([]byte, error) {
	endpoint := strings.Join([]string{*HttpServer, "query"}, "/")
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

func Proxy(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't read body", http.StatusBadRequest)
		return
	}

	// TODO - Later send bytes directly to Dgraph.
	res, err := Query(string(b))
	if err != nil {
		sr.Write(w, err.Error(), "Couldn't read body", http.StatusBadRequest)
		return
	}
	w.Write(res)
}
