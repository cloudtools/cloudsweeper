// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
)

const (
	smtpServer    = "smtp.gmail.com"
	smtpPort      = 587
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
	displayName string
}

// NewClient will create a new email client for sending mails
func NewClient(username, password, displayName string) Client {
	auth := smtp.PlainAuth("", username, password, smtpServer)
	m := new(mailer)
	m.auth = auth
	m.user = username
	m.displayName = displayName

	return m
}

// SendEmail will send a mail to the specified address. Please note that
// the content is not HTML escaped. That would be up to whoever uses the method
func (m *mailer) SendEmail(subject, content string, recipients ...string) error {
	server := fmt.Sprintf("%s:%d", smtpServer, smtpPort)
	var msg bytes.Buffer

	context := &mailContext{
		From:        m.user,
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

	err = smtp.SendMail(server, m.auth, m.user, recipients, msg.Bytes())
	return err
}

type mailContext struct {
	From        string
	To          string
	Subject     string
	Body        string
	DisplayName string
}
