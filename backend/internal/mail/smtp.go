package mail

import (
	"context"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
)

// RequestIDSubjectPrefix は管理者宛の承認依頼メールの件名に付与し、
// IMAPポーラー(poller.go)が返信を検知する際の照合に使う。
const RequestIDSubjectPrefix = "[しりとりアプリ] パスワード再設定の承認依頼 (ID: "

type SMTPMailer struct {
	host       string
	port       int
	username   string
	password   string
	from       string
	adminEmail string
}

func NewSMTPMailer(host string, port int, username, password, from, adminEmail string) *SMTPMailer {
	return &SMTPMailer{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		from:       from,
		adminEmail: adminEmail,
	}
}

func (m *SMTPMailer) SendAdminApprovalRequest(ctx context.Context, requestID, userEmail string) error {
	subject := fmt.Sprintf("%s%s)", RequestIDSubjectPrefix, requestID)
	body := fmt.Sprintf(
		"ユーザー(%s)からパスワード再設定の申請がありました。\n\n"+
			"承認する場合は、このメールに本文を空にしたまま返信してください。\n"+
			"内容に心当たりがない場合は、返信せず破棄してください。\n",
		userEmail,
	)
	return m.send(m.adminEmail, subject, body)
}

func (m *SMTPMailer) SendNewPassword(ctx context.Context, userEmail, newPassword string) error {
	subject := "しりとりアプリ: 新しいパスワードのお知らせ"
	body := fmt.Sprintf(
		"パスワード再設定の申請が承認されました。\n\n"+
			"新しいパスワード: %s\n\n"+
			"このパスワードでログインし、必要に応じて改めてパスワードを変更してください。\n",
		newPassword,
	)
	return m.send(userEmail, subject, body)
}

func (m *SMTPMailer) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	crlfBody := strings.ReplaceAll(body, "\n", "\r\n")
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, to, encodedSubject, crlfBody)
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}
