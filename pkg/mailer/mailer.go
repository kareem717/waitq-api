package mailer

import (
	resendMailer "waitq/api/pkg/mailer/resend"

	"github.com/resend/resend-go/v2"
)

type ResendMailer interface {
	SendUnsubscribeRequestEmail(waitlistName string, unsubscribeLink string, email string) error
}

type Mailer struct {
	ResendMailer ResendMailer
}

func NewMailer(
	resendClient *resend.Client,
) *Mailer {
	return &Mailer{
		ResendMailer: resendMailer.New(resendClient),
	}
}
