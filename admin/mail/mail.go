package mail

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/dgraph-io/gru/admin/company"
	"github.com/dgraph-io/gru/x"
	"github.com/russross/blackfriday"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SENDGRID_API_KEY = flag.String("sendgrid", "", "Sendgrid API Key")

// TODO - Later just have one IP address with port info.
var Ip = flag.String("ip", "https://gru.dgraph.io", "Public IP address of server")

func GetInviteUrl(token string) string {
	return fmt.Sprintf("%v/#/quiz/%v", *Ip, token)
}

func Send(email, validity, token string) {
	if *SENDGRID_API_KEY == "" {
		fmt.Println(*Ip + "/#/quiz/" + token)
		return
	}
	fmt.Println("Sending invite: " + *Ip + "/#/quiz/" + token)

	c, err := company.Info()
	if err != nil {
		fmt.Println("Invite not sent", err)
		return
	}

	from := mail.NewEmail(c.Name, c.Email)
	subject := fmt.Sprintf("Invitation for screening quiz from %v", c.Name)
	to := mail.NewEmail("", email)
	URL := GetInviteUrl(token)
	// Lets unescape it first.
	invite, err := url.QueryUnescape(c.Invite)
	if err != nil {
		fmt.Println("Invite not sent", err)
		return
	}

	hr := blackfriday.HtmlRenderer(0, "", "")
	o := blackfriday.Options{}
	o.Extensions = blackfriday.EXTENSION_HARD_LINE_BREAK
	invite = string(blackfriday.MarkdownOptions([]byte(invite), hr, o))

	body := `
<html>
<body>
Hello!
<br/><br/>
You have been invited to take the screening quiz by ` + c.Name + `.
<br/><br/>
You can take the quiz any time till ` + validity + ` by visiting <a href="` + URL + `"> ` + URL + `</a>.
<br/>
` + invite + `
</body>
</html>
`
	content := mail.NewContent("text/html", body)
	m := mail.NewV3MailInit(from, subject, to, content)
	request := sendgrid.GetRequest(*SENDGRID_API_KEY, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)

	x.Debug("Sending invite to " + email + "\n  Invite URL is " + URL)

	_, err = sendgrid.API(request)
	if err != nil {
		fmt.Println("Invite not sent", err)
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
