package mail

import (
	"flag"
	"fmt"

	"github.com/dgraph-io/gru/x"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SENDGRID_API_KEY = flag.String("sendgrid", "", "Sendgrid API Key")

// TODO - Later just have one IP address with port info.
var ip = flag.String("ip", "http://localhost", "Public IP address of server")

func Send(name, email, token string) {
	if *SENDGRID_API_KEY == "" {
		fmt.Println("localhost:8000/#/quiz/" + token)
		return
	}
	from := mail.NewEmail("Dgraph", "join@dgraph.io")
	subject := "Invitation for screening quiz from Dgraph"
	to := mail.NewEmail(name, email)
	// TODO - Move this to a template.
	url := fmt.Sprintf("%v:8000/#/quiz/%v", *ip, token)
	body := `
<html>
<head>
    <title></title>
</head>
<body>
Hello ` + name + `,
<br /><br/>
You have been invited to take the screening quiz by Dgraph. You can take the quiz anytime by
visiting <a href="` + url + `" target="_blank">` + url + `</a>.
<br /><br/>
</body>
</html>
`
	content := mail.NewContent("text/html", body)
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

func SendReport(name string, score, maxScore float64, body string) {
	if *SENDGRID_API_KEY == "" {
		return
	}

	from := mail.NewEmail("Dgraph", "join@dgraph.io")
	subject := fmt.Sprintf("Gru: Candidate %v scored %.1f/%.1f in the demo test", name,
		score, maxScore)
	to := mail.NewEmail("Dgraph", "join@dgraph.io")
	content := mail.NewContent("text/html", body)
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
