package resend

import (
	"bytes"
	"html/template"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

type ResendMailer struct {
	client *resend.Client
	logger *zap.Logger
}

func New(client *resend.Client) *ResendMailer {
	return &ResendMailer{
		client: client,
	}
}

func (rm *ResendMailer) SendUnsubscribeRequestEmail(waitlistName string, unsubscribeLink string, email string) error {
	t, err := template.ParseFiles("./templates/unsubscribe_request.html")
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, struct {
		WaitlistName string
		UnsubscribeLink string
	}{
		WaitlistName: waitlistName,
		UnsubscribeLink: unsubscribeLink,
	}); err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		To:      []string{email},
		From:    "WaitQ <waitq@waitq.com>",
		Html:    tpl.String(),
		Subject: "Unsubscribe Request",
	}

	sent, err := rm.client.Emails.Send(params)
	if err != nil {
		rm.logger.Error("failed to send unsubscribe request email", zap.Error(err))
		return err
	}

	rm.logger.Info("unsubscribe request email sent", zap.Any("sentId", sent.Id))

	return nil
}
