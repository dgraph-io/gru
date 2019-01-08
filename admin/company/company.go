package company

import (
	"fmt"

	"github.com/dgraph-io/gru/dgraph"
)

type Company struct {
	Name        string `json:"company.name"`
	Email       string `json:"company.email"`
	Backup      int    `json:"backup,string"`
	BackupDays  int    `json:"backup_days,string"`
	Invite      string `json:"company.invite_email"`
	RejectEmail string `json:"company.reject_email"`
	Reject      bool   `json:"company.reject,string"`
}

type info struct {
	Companies []Company `json:"info"`
}

func Info() (Company, error) {
	q := `{
	    info(func: has(is_company_info)) {
            company.name
            company.email
            company.reject
            company.invite_email
            company.reject_email
            backup
            backup_days
        }
    }`

	var companies info
	if err := dgraph.QueryAndUnmarshal(q, &companies); err != nil {
		return Company{}, err
	}

	if len(companies.Companies) != 1 {
		return Company{}, fmt.Errorf("No company information found.")
	}

	com := companies.Companies[0]
	return com, nil
}
