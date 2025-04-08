package mail

import (
	"context"
	"log"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type Email struct {
	Subject string
	Body    string
	From    string
	To      []string
}

type Mailer interface {
	SendMail(e *Email)
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

func (m *Mailgun) SendMail(e *Email) {
	mg := mailgun.NewMailgun(m.domain, m.apiKey)
	mg.SetAPIBase(m.apiBase)

	message := mailgun.NewMessage(e.From, e.Subject, e.Body, e.To...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, _, err := mg.Send(ctx, message)
	if err != nil {
		log.Fatal(err)
	}
}
