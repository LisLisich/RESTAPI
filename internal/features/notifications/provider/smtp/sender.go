package smtp

import (
	"context"
	"fmt"
	"mime"
	netsmtp "net/smtp"
	"strings"
)

type Sender struct {
	address  string
	host     string
	username string
	password string
	from     string
}

func NewSender(config Config) *Sender {
	return &Sender{
		address:  config.Address,
		host:     config.Host,
		username: config.Username,
		password: config.Password,
		from:     config.From,
	}
}

func (sender *Sender) Send(
	ctx context.Context,
	to string,
	subject string,
	body string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message, err := buildMessage(sender.from, to, subject, body)
	if err != nil {
		return err
	}
	auth := netsmtp.PlainAuth("", sender.username, sender.password, sender.host)
	if err := netsmtp.SendMail(sender.address, auth, sender.from, []string{to}, message); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}
	return nil
}

type DisabledSender struct{}

func (DisabledSender) Send(context.Context, string, string, string) error {
	return nil
}

func buildMessage(from string, to string, subject string, body string) ([]byte, error) {
	for name, value := range map[string]string{
		"from": from, "to": to, "subject": subject,
	} {
		if strings.ContainsAny(value, "\r\n") {
			return nil, fmt.Errorf("%s contains a newline", name)
		}
	}
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}
	return []byte(strings.Join(headers, "\r\n")), nil
}
