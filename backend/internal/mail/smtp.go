package mail

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// RequestIDSubjectPrefix は管理者宛の承認依頼メールの件名に付与し、
// IMAPポーラー(poller.go)が返信を検知する際の照合に使う。
const RequestIDSubjectPrefix = "[しりとりアプリ] パスワード再設定の承認依頼 (ID: "

// smtpTimeout はSMTP接続〜送信完了までに許容する最大時間。
// ctx にこれより早いデッドラインが設定されていればそちらを優先する。
// net/smtp.SendMail 自体にはタイムアウトの仕組みがなく、接続先が応答しない場合に
// 呼び出し元(HTTPハンドラ)を無期限にブロックしてしまうため、独自に上限を設けている。
const smtpTimeout = 10 * time.Second

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
	return m.send(ctx, m.adminEmail, subject, body)
}

func (m *SMTPMailer) SendNewPassword(ctx context.Context, userEmail, newPassword string) error {
	subject := "しりとりアプリ: 新しいパスワードのお知らせ"
	body := fmt.Sprintf(
		"パスワード再設定の申請が承認されました。\n\n"+
			"新しいパスワード: %s\n\n"+
			"このパスワードでログインし、必要に応じて改めてパスワードを変更してください。\n",
		newPassword,
	)
	return m.send(ctx, userEmail, subject, body)
}

func (m *SMTPMailer) send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	crlfBody := strings.ReplaceAll(body, "\n", "\r\n")
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, to, encodedSubject, crlfBody)

	deadline := time.Now().Add(smtpTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	dialer := net.Dialer{Deadline: deadline}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	defer conn.Close()
	// 接続確立後の全I/O(EHLO/AUTH/DATA等)にも同じ上限を適用する。
	// これにより、相手が応答を返さない(パケットを黙って捨てる)場合でも
	// 呼び出し元は最大 smtpTimeout 秒でエラーを受け取れる。
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("smtp set deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", m.username, m.password, m.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(m.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close body: %w", err)
	}
	return client.Quit()
}
