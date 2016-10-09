package mail

import (
	"flag"
	"fmt"

	"github.com/dgraph-io/gru/x"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SENDGRID_API_KEY = flag.String("sendgrid", "", "Sendgrid API Key")

func Send(name, email, token string) {
	if *SENDGRID_API_KEY == "" {
		return
	}
	from := mail.NewEmail("Dgraph", "join@dgraph.io")
	subject := "Invitation for screening quiz from Dgraph"
	to := mail.NewEmail(name, email)
	// TODO - Invite formatting of the mail, make the link an actual link.
	content := mail.NewContent("text/plain", `
    You have been invited to take the screening quiz by Dgraph. You can take the quiz anytime by
	visiting localhost:8082/quiz/`+token+`.`)
	m := mail.NewV3MailInit(from, subject, to, content)

	request := sendgrid.GetRequest(*SENDGRID_API_KEY, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	}
	x.Debug("Mail sent")
	x.Debug(response.StatusCode)
	x.Debug(response.Body)
	x.Debug(response.Headers)
}
