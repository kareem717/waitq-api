package stripe

import (
	"net/url"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

type ClientConfig struct {
	StripeAPIKey             string
	BaseURL                  string
	SubscriptionCallbackPath string
}

type Client struct {
	config ClientConfig
}

func NewClient(config ClientConfig) *Client {
	// parse baseURL
	_, err := url.Parse(config.BaseURL)
	if err != nil {
		panic("An error occurred while parsing the base URL: " + err.Error() + "\nBase URL: " + config.BaseURL)
	}

	return &Client{config: config}
}

func (c *Client) CreateCheckoutSession(priceID string, customerAccountId string) (*stripe.CheckoutSession, error) {
	stripe.Key = c.config.StripeAPIKey
	checkoutParams := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(c.config.BaseURL + c.config.SubscriptionCallbackPath + "?session_id={CHECKOUT_SESSION_ID}"),
		Metadata: map[string]string{
			"customer_account_id": customerAccountId,
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
	}

	s, err := session.New(checkoutParams)
	return s, err
}

func (c *Client) GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	stripe.Key = c.config.StripeAPIKey
	sess, err := session.Get(sessionID, nil)
	return sess, err
}

func (c *Client) GetSessionLineItemIter(sess *stripe.CheckoutSession) (*session.LineItemIter, error) {
	lineItemParams := stripe.CheckoutSessionListLineItemsParams{}
	lineItemParams.Session = stripe.String(sess.ID)
	iter := session.ListLineItems(&lineItemParams)
	if !iter.Next() {
		return nil, iter.Err()
	}
	
	return iter, nil
}
