// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

// Package mailer is a utility to send email. Configuration is not within
// the scope of this package, it simply takes an SMTP server, port,
// username and password as an argument to the NewClient function.
//
// This has been tested with Gmail using smtp.gmail.com and port 587
package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
)

const (
	emailTemplate = `From: {{ .DisplayName }} <{{- .From -}}>
To: {{ .To }}
Subject: {{ .Subject }}
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";

{{ .Body }}`
)

// Client is used to send emails using standard settings
type Client interface {
	// SendEmail will send a mail to the specified email address
	SendEmail(subject, content string, recipients ...string) error
}

type mailer struct {
	user        string
	auth        smtp.Auth
	from        string
	displayName string
	smtpServer  string
	smtpPort    int
}

// NewClient will create a new email client for sending mails
func NewClient(username, password, displayName, from, smtpServer string, smtpPort int) Client {
	auth := smtp.PlainAuth("", username, password, smtpServer)
	m := new(mailer)
	m.auth = auth
	m.from = from
	m.displayName = displayName
	m.smtpServer = smtpServer
	m.smtpPort = smtpPort

	return m
}

// SendEmail will send a mail to the specified address. Please note that
// the content is not HTML escaped. That would be up to whoever uses the method
func (m *mailer) SendEmail(subject, content string, recipients ...string) error {
	server := fmt.Sprintf("%s:%d", m.smtpServer, m.smtpPort)
	var msg bytes.Buffer

	context := &mailContext{
		From:        m.from,
		To:          strings.Join(recipients, ", "),
		Subject:     subject,
		Body:        content,
		DisplayName: m.displayName,
	}

	t := template.New("mailTemplate")
	t, err := t.Parse(emailTemplate)
	if err != nil {
		return err
	}

	err = t.Execute(&msg, context)
	if err != nil {
		return err
	}

	err = smtp.SendMail(server, m.auth, m.from, recipients, msg.Bytes())
	return err
}

type mailContext struct {
	From        string
	To          string
	Subject     string
	Body        string
	DisplayName string
}
