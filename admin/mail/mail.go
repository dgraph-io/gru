package mail

import (
	"flag"
	"fmt"

	"github.com/dgraph-io/gru/x"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SENDGRID_API_KEY = flag.String("sendgrid", "", "Sendgrid API Key")
var reportMail = flag.String("report", "join@dgraph.io", "Email on which to send the reports.")

// TODO - Later just have one IP address with port info.
var Ip = flag.String("ip", "http://localhost:2020", "Public IP address of server")

func Send(email, validity, token string) {
	if *SENDGRID_API_KEY == "" {
		fmt.Println(*Ip + "/#/quiz/" + token)
		return
	}
	from := mail.NewEmail("Dgraph", "join@dgraph.io")
	subject := "Invitation for screening quiz from Dgraph"
	to := mail.NewEmail("", email)
	// TODO - Move this to a template.
	url := fmt.Sprintf("%v/#/quiz/%v", *Ip, token)
	body := `
<html>
<head>
    <title></title>
</head>
<body>
Hello!
<br/><br/>
You have been invited to take the screening quiz by Dgraph.
<br/>
You can take the quiz anytime till ` + validity + ` by visiting <a href="` + url + `" target="_blank">` + url + `</a>.
<br/>
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
		return
	}
	x.Debug("Mail sent")
	x.Debug(response.StatusCode)
	x.Debug(response.Body)
	x.Debug(response.Headers)
}

func SendReport(name string, quiz string, score, maxScore float64, body string) {
	if *SENDGRID_API_KEY == "" {
		return
	}

	from := mail.NewEmail("Gru", "join@dgraph.io")
	subject := fmt.Sprintf("%v scored %.2f/%.2f in the %v quiz", name,
		score, maxScore, quiz)
	to := mail.NewEmail("Dgraph", *reportMail)

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
