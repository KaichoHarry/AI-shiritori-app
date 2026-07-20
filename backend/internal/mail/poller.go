package mail

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// PasswordResetApprover はIMAPポーラーが管理者の空メール返信を検知した際に呼び出す。
// 実装は internal/auth.Service が満たす。
type PasswordResetApprover interface {
	ApprovePasswordReset(ctx context.Context, requestID string) error
}

var requestIDPattern = regexp.MustCompile(`\(ID: ([0-9a-fA-F-]{36})\)`)

// processedFlag はIMAPの標準\Seenフラグではなく、本アプリ専用のカスタムキーワードとして
// 処理済みメールを管理する。Gmailはスレッド表示の都合で返信メール自体を自動的に既読
// (\Seen)にすることがあり、\Seenに依存すると未処理の返信を見逃す恐れがあるため。
const processedFlag = "ShiritoriProcessed"

type IMAPPoller struct {
	host         string
	port         int
	username     string
	password     string
	pollInterval time.Duration
	approver     PasswordResetApprover
}

func NewIMAPPoller(host string, port int, username, password string, pollInterval time.Duration, approver PasswordResetApprover) *IMAPPoller {
	return &IMAPPoller{
		host:         host,
		port:         port,
		username:     username,
		password:     password,
		pollInterval: pollInterval,
		approver:     approver,
	}
}

// Run はcontextがキャンセルされるまで、pollIntervalごとに管理者メールボックスを
// ポーリングし続ける。呼び出し元がgoroutineとして起動する想定。
func (p *IMAPPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		if err := p.pollOnce(ctx); err != nil {
			log.Printf("imap poller: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *IMAPPoller) pollOnce(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", p.host, p.port)
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("dial imap: %w", err)
	}
	defer c.Logout()

	if err := c.Login(p.username, p.password); err != nil {
		return fmt.Errorf("imap login: %w", err)
	}

	if _, err := c.Select("INBOX", false); err != nil {
		return fmt.Errorf("select inbox: %w", err)
	}

	criteria := &imap.SearchCriteria{
		Header:       textproto.MIMEHeader{"Subject": []string{RequestIDSubjectPrefix}},
		WithoutFlags: []string{processedFlag},
	}
	uids, err := c.UidSearch(criteria)
	if err != nil {
		return fmt.Errorf("uid search: %w", err)
	}
	if len(uids) == 0 {
		return nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, section.FetchItem()}

	messages := make(chan *imap.Message, len(uids))
	fetchErr := make(chan error, 1)
	go func() {
		fetchErr <- c.UidFetch(seqSet, items, messages)
	}()

	for msg := range messages {
		if err := p.handleMessage(ctx, msg, section); err != nil {
			log.Printf("imap poller: handle message uid=%v: %v", msg.Uid, err)
		}
	}
	if err := <-fetchErr; err != nil {
		return fmt.Errorf("uid fetch: %w", err)
	}

	// 承認済み・却下いずれの場合も処理済みフラグを付け、次回ポーリングでの再処理を防ぐ。
	markSeq := new(imap.SeqSet)
	markSeq.AddNum(uids...)
	storeItem := imap.FormatFlagsOp(imap.AddFlags, true)
	if err := c.UidStore(markSeq, storeItem, []interface{}{processedFlag}, nil); err != nil {
		log.Printf("imap poller: mark processed: %v", err)
	}

	return nil
}

func (p *IMAPPoller) handleMessage(ctx context.Context, msg *imap.Message, section *imap.BodySectionName) error {
	if msg.Envelope == nil {
		return nil
	}

	match := requestIDPattern.FindStringSubmatch(msg.Envelope.Subject)
	if match == nil {
		return nil
	}
	requestID := match[1]

	literal := msg.GetBody(section)
	if literal == nil {
		return nil
	}
	plainText, err := extractPlainText(literal)
	if err != nil {
		return fmt.Errorf("extract plain text: %w", err)
	}

	if !isEmptyReply(plainText) {
		return nil
	}

	return p.approver.ApprovePasswordReset(ctx, requestID)
}

// extractPlainText はメッセージ全体(RFC822形式)のio.Readerから、multipart/alternative等の
// MIME構造を解釈してtext/plain部分のデコード済みテキストを取り出す。
func extractPlainText(r io.Reader) ([]byte, error) {
	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, err
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		h, ok := p.Header.(*mail.InlineHeader)
		if !ok {
			continue
		}
		contentType, _, err := h.ContentType()
		if err != nil || !strings.HasPrefix(contentType, "text/plain") {
			continue
		}
		return io.ReadAll(p.Body)
	}

	return nil, nil
}

// isEmptyReply は返信本文が実質的に空(引用文・署名以外に本文がない)かどうかを判定する。
//
// Gmail等のメールクライアントは返信時に「YYYY年M月D日(曜) H:mm 差出人 <email>:」のような
// 引用元の attribution 行に続けて、元メール全文を "> " 付きの引用ブロックとして自動的に
// 本文へ追加する。そのため単純に空白行・"> "行を除去するだけでは不十分で、attribution行
// (多くのメールクライアントで ":" で終わる)が現れた時点でそれ以降を引用とみなして無視する。
//
// TODO: これはヒューリスティックであり、メールクライアントの言語設定やフォーマットに
// よっては attribution行を誤検知/see過ごす可能性がある。より厳密には引用元メールの
// Message-IDから元本文と突き合わせて差分を取る等の改善余地がある。
func isEmptyReply(body []byte) bool {
	text := string(body)
	var kept []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, ">") {
			break
		}
		if isQuoteAttributionLine(trimmed) {
			break
		}
		kept = append(kept, trimmed)
	}
	return len(kept) == 0
}

func isQuoteAttributionLine(line string) bool {
	if len(line) < 10 {
		return false
	}
	return strings.HasSuffix(line, ":") ||
		strings.Contains(line, "wrote:") ||
		strings.Contains(line, "さんは書きました")
}
