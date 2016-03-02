package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/mail"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockSender struct {
	addr string
	auth smtp.Auth
	from string
	to   []string
	msg  string
}

func (m *mockSender) MockSender(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.addr = addr
	m.auth = a
	m.from = from
	m.to = to
	m.msg = string(msg)

	return nil
}

func TestSendEmail(t *testing.T) {
	assert := assert.New(t)

	var m mockSender

	smtpAuth := smtpAuth{Username: "username", Password: "password"}
	cfg := smtpConfiguration{
		Host: "host",
		Port: 6666,
		Auth: &smtpAuth,
		From: "from",
		To:   []string{"to1", "to2"}}

	errorStrings := []string{"Error1"}

	const rfc2822 = "Mon, 02 Jan 2006 15:04:05 -0700"
	tt := "Sun, 28 Feb 2016 18:54:05 +0100"
	parsedTime, err := time.Parse(rfc2822, tt)
	assert.Nil(err, fmt.Sprint(err))

	err = sendEmail(m.MockSender, cfg, parsedTime, errorStrings)
	assert.Nil(err, fmt.Sprint(err))

	assert.Equal("host:6666", m.addr)
	assert.Equal("from", m.from)
	assert.Equal(2, len(m.to))
	assert.Equal("to1", m.to[0])
	//assert.Equal("to2", m.to[1])
	assert.NotNil(m.auth)

	r := strings.NewReader(m.msg)
	mm, err := mail.ReadMessage(r)
	assert.Nil(err, fmt.Sprint(err))

	header := mm.Header
	assert.Equal("Sun, 28 Feb 2016 18:54:05 +0100", header.Get("Date"))
	assert.Equal("<from@>", header.Get("From"))
	assert.Equal("<to1@>, <to2@>", header.Get("To"))
	assert.Equal("Ismonitor alert", header.Get("Subject"))
	assert.Equal("1.0", header.Get("MIME-Version"))
	assert.Equal("text/plain; charset=\"utf-8\"", header.Get("Content-Type"))
	assert.Equal("base64", header.Get("Content-Transfer-Encoding"))

	body, err := ioutil.ReadAll(mm.Body)
	assert.Nil(err)
	// decode the base64 encoded body
	decodedStr := make([]byte, base64.StdEncoding.DecodedLen(len(body)))
	len, err := base64.StdEncoding.Decode(decodedStr, body)
	assert.Nil(err)
	// DecodedLen returns the maximal length.
	// This length is useful for sizing your buffer but part of the buffer won't be
	// written and thus won't be valid UTF-8.
	// You have to use only the real written length returned by the Decode function.
	// Hence the [:len] below
	assert.Equal("Error1\n", string(decodedStr[:len]))
}
