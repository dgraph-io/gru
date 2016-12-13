package company

import (
	"fmt"

	"github.com/dgraph-io/gru/dgraph"
)

type Company struct {
	Name  string `json:"company.name"`
	Email string `json:"company.email"`
}

type info struct {
	Companies []Company `json:"info"`
}

func Info() (Company, error) {
	c := Company{Name: "Dgraph",
		Email: "join@dgraph.io",
	}

	q := `{
	    info(_xid_: root) {
            company.name
            company.email
        }
    }`

	var companies info
	if err := dgraph.QueryAndUnmarshal(q, &companies); err != nil {
		return c, err
	}

	if len(companies.Companies) != 1 {
		return c, fmt.Errorf("No company information found.")
	}

	com := companies.Companies[0]
	return Company{
		Name:  com.Name,
		Email: com.Email,
	}, nil
}
