package main

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"
)

type mailSender func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

func sendEmail(senderFunc mailSender, smtpConfig smtpConfiguration, ts time.Time, errors []string) error {
	// set up possible authentication
	var auth smtp.Auth
	if smtpConfig.Auth != nil {
		a := *smtpConfig.Auth
		auth = smtp.PlainAuth(
			"",
			a.Username,
			a.Password,
			smtpConfig.Host,
		)
	} else {
		auth = nil
	}

	from := mail.Address{Address: smtpConfig.From}
	var toAddresses []mail.Address
	var toString string
	for i, t := range smtpConfig.To {
		ma := mail.Address{Address: t}
		toAddresses = append(toAddresses, ma)
		toString += ma.String()
		if i != len(smtpConfig.To)-1 {
			// add comma between addresses unless it's the last one
			toString += ", "
		}
	}
	title := "Ismonitor alert"

	body := ""
	for _, e := range errors {
		body += fmt.Sprintf("%s\n", e)
	}

	const rfc2822 = "Mon, 02 Jan 2006 15:04:05 -0700"

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = toString
	header["Subject"] = title
	header["Date"] = ts.Format(rfc2822)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	return senderFunc(
		smtpConfig.Host+":"+fmt.Sprintf("%d", smtpConfig.Port),
		auth,
		from.Address,
		smtpConfig.To,
		[]byte(message))
}
