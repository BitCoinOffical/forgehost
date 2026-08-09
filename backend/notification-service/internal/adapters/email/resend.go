package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/resend/resend-go/v2"
)

const (
	resendFrom = "ForgeHost <onboarding@resend.dev>"
)

type ResendConfig struct {
	ResendApiKey string
}

type ResendClient struct {
	client *resend.Client
}

type Data struct {
	Code string
}

func NewResendSender(cfg *ResendConfig) *ResendClient {
	client := resend.NewClient(cfg.ResendApiKey)
	return &ResendClient{client: client}
}

//go:embed template/code.html
var templateFS embed.FS

func (g *ResendClient) SendCodeEmail(to []string, code string, Subject string) (string, error) {

	tmpl, err := template.ParseFS(templateFS, "template/code.html")
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
		Subject: Subject,
		Html:    buf.String(),
	}

	sent, err := g.client.Emails.Send(params)
	if err != nil {
		return "", fmt.Errorf("g.client.Emails.Send: %w", err)
	}

	return sent.Id, nil
}
