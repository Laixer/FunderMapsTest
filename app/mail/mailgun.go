package mail

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type Email struct {
	Subject      string
	Body         string
	From         string
	To           []string
	Template     string
	TemplateVars map[string]any
}

type Mailer interface {
	SendMail(e *Email) error
}

type Mailgun struct {
	domain  string
	apiKey  string
	apiBase string
}

func NewMailer(domain, apiKey, apiBase string) *Mailgun {
	return &Mailgun{
		domain:  domain,
		apiKey:  apiKey,
		apiBase: apiBase,
	}
}

func (m *Mailgun) SendMail(e *Email) error {
	mg := mailgun.NewMailgun(m.domain, m.apiKey)
	mg.SetAPIBase(m.apiBase)

	message := mailgun.NewMessage(e.From, e.Subject, e.Body, e.To...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, _, err := mg.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mailgun) SendTemplatedMail(e *Email) error {
	mg := mailgun.NewMailgun(m.domain, m.apiKey)
	mg.SetAPIBase(m.apiBase)

	message := mailgun.NewMessage(e.From, e.Subject, "", e.To...)
	message.SetTemplate(e.Template)

	if e.TemplateVars != nil {
		for k, v := range e.TemplateVars {
			message.AddTemplateVariable(k, v)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := mg.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
