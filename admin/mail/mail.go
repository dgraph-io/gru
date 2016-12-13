package mail

import (
	"flag"
	"fmt"

	"github.com/dgraph-io/gru/admin/company"
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

	c, err := company.Info()
	if err != nil {
		fmt.Println(err)
		return
	}

	from := mail.NewEmail(c.Name, c.Email)
	subject := fmt.Sprintf("Invitation for screening quiz from %v", c.Name)
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
You have been invited to take the screening quiz by ` + c.Name + `.
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
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	x.Debug("Mail sent")
}

func SendReport(name string, quiz string, score, maxScore float64, body string) {
	if *SENDGRID_API_KEY == "" {
		return
	}

	c, err := company.Info()
	if err != nil {
		fmt.Println(err)
		return
	}

	from := mail.NewEmail("Gru", c.Email)
	subject := fmt.Sprintf("%v scored %.2f/%.2f in the %v quiz", name,
		score, maxScore, quiz)
	to := mail.NewEmail(c.Name, c.Email)

	content := mail.NewContent("text/html", body)
	m := mail.NewV3MailInit(from, subject, to, content)
	request := sendgrid.GetRequest(*SENDGRID_API_KEY, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	}
	x.Debug("Mail sent")
}

func Reject(name, email string) {
	if *SENDGRID_API_KEY == "" {
		fmt.Printf("Sending rejection mail to %v\n", name)
		return
	}
	from := mail.NewEmail("Pulkit Jain", "pulkit@dgraph.io")
	subject := "Dgraph <> Quiz"
	p := mail.NewPersonalization()
	to := mail.NewEmail(name, email)
	p.AddTos(to)
	cc := mail.NewEmail("Dgraph", *reportMail)
	p.AddCCs(cc)
	body := `
<html>
<head>
    <title></title>
</head>
<body>
Hi ` + name + `,
<br/><br/>
Thanks for taking the time to complete the quiz. Unfortunately, the quiz score didn’t meet the expectations we had, so we’ve decided not to move forward with discussions regarding the full-time role.
<br/><br/>
Good luck with your future endeavors. If Dgraph interests you, I will encourage you to contribute to Dgraph as an open source contributor. A good starting point is <a href="https://dgraph.io">dgraph.io</a> where you’ll find links to our Slack and Discourse channels where we hang out.
<br/><br/>
Thanks<br/>
Pulkit Rai<br/>
<a href="https://dgraph.io">https://dgraph.io</a><br/>
</body>
</html>
`
	content := mail.NewContent("text/html", body)
	m := mail.NewV3MailInit(from, subject, to, content)
	m.AddPersonalizations(p)
	request := sendgrid.GetRequest(*SENDGRID_API_KEY, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	_, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
		return
	}
}
