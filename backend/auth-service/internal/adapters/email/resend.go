package email

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/resend/resend-go/v2"
)

const (
	resendFrom = "onboarding@resend.dev"
)

type ResendConfig struct {
	ResendApiKey string
}

type ResendClient struct {
	client *resend.Client
}

type Data struct {
	Code int
}

func NewResendSender(cfg *ResendConfig) *ResendClient {
	client := resend.NewClient(cfg.ResendApiKey)
	return &ResendClient{client: client}
}

func (g *ResendClient) SendVerificationEmail(to []string, code int) (string, error) {
	tmpl, err := template.ParseFiles("template/code.html")
	if err != nil {
		return "", fmt.Errorf("template.ParseFiles: %w", err)
	}
	var buf bytes.Buffer
	tmpl.Execute(&buf, Data{
		Code: code,
	})

	params := &resend.SendEmailRequest{
		From:    resendFrom,
		To:      to,
		Subject: "Verification Code",
		Html:    buf.String(),
	}

	sent, err := g.client.Emails.Send(params)
	if err != nil {
		return "", fmt.Errorf("g.client.Emails.Send: %w", err)
	}

	return sent.Id, nil
}
