package main

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"
)

type mailSender func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

func sendEmail(senderFunc mailSender, smtpConfig smtpConfiguration, ts time.Time, errors []verificationError) error {
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
	toString := makeToAddresses(smtpConfig.To)

	title := "Ismonitor alert"

	body := makeMessage(errors)

	message := makeHeaders(from.String(), toString, title, ts) // headers
	message += "\r\n"
	message += base64.StdEncoding.EncodeToString([]byte(body)) // body

	return senderFunc(
		smtpConfig.Host+":"+fmt.Sprintf("%d", smtpConfig.Port),
		auth,
		from.Address,
		smtpConfig.To,
		[]byte(message))
}

func makeToAddresses(to []string) string {
	var toString string
	for i, t := range to {
		ma := mail.Address{Address: t}
		toString += ma.String()
		if i != len(to)-1 {
			// add comma between addresses unless it's the last one
			toString += ", "
		}
	}
	return toString
}

func makeHeaders(from string, to string, title string, ts time.Time) string {
	const rfc2822 = "Mon, 02 Jan 2006 15:04:05 -0700"

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = title
	header["Date"] = ts.Format(rfc2822)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	return message
}

func makeMessage(errors []verificationError) string {
	body := ""
	for _, e := range errors {
		body += fmt.Sprintf("%s\n   %s\n", e.title, e.message)
	}

	return body
}
