package quiz

import (
	"fmt"
	"time"

	"github.com/dgraph-io/gru/admin/mail"
	"github.com/dgraph-io/gru/dgraph"
)

type rejectRes struct {
	Root []struct {
		Candidates []cand `json:"candidate"`
	} `json:"root"`
}

func reject() error {
	q := `
        {
                root(_xid_:rejected) {
                        candidate {
                                _uid_
                                name
                                email
                                completed_at
                        }
                }
        }`

	var resp rejectRes
	if err := dgraph.QueryAndUnmarshal(q, &resp); err != nil {
		return err
	}
	if len(resp.Root) != 1 {
		return nil
	}

	now := time.Now()
	gap := 2 * time.Hour
	cands := resp.Root[0].Candidates
	for _, c := range cands {
		if now.Before(c.CompletedAt.Add(gap)) {
			// We don't include this candidate in the candidates that we should
			// be rejecting now.
			continue
		}
		mail.Reject(c.Name, c.Email)
		// Lets delete this node now.
		m := new(dgraph.Mutation)
		m.Del(`<rejected> <candidate> <_uid_:` + c.Id + `> .`)
		if _, err := dgraph.SendMutation(m.String()); err != nil {
			return err
		}

	}
	return nil
}

func Reject() {
	t := time.Tick(10 * time.Minute)
	for range t {
		if err := reject(); err != nil {
			fmt.Printf("Error while rejecting candidates: %v", err)
		}
	}
}
