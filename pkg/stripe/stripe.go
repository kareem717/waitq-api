package stripe

import (
	"errors"
	"net/url"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/subscription"
)

type ClientConfig struct {
	StripeAPIKey                string
	BaseURL                     string
	SubscriptionCallbackPath    string
	StripeProProductID          string
	StripeEntrepreneurProductID string
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

func (c *Client) CreateCheckoutSession(priceID string, customerAccountId string, redirectUrl string) (*stripe.CheckoutSession, error) {
	stripe.Key = c.config.StripeAPIKey
	checkoutParams := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(c.config.BaseURL + c.config.SubscriptionCallbackPath + "?session_id={CHECKOUT_SESSION_ID}"),
		Metadata: map[string]string{
			"customer_account_id": customerAccountId,
			"redirect_url":        redirectUrl,
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

func (c *Client) GetSession(sessionID string) (*stripe.CheckoutSession, error) {
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

func (c *Client) GetCustomerSubscriptions(customerID string) (*subscription.Iter, error) {
	stripe.Key = c.config.StripeAPIKey

	params := &stripe.SubscriptionListParams{Customer: stripe.String(customerID)}
	result := subscription.List(params)

	return result, nil
}

func (c *Client) UpdateCustomerSubscription(subscriptionID string, newPriceID string) (*stripe.Subscription, error) {
	stripe.Key = c.config.StripeAPIKey

	// Retrieve the subscription
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	if len(sub.Items.Data) == 0 {
		return nil, errors.New("no subscription items found")
	}

	// Use the subscription item ID
	subscriptionItemID := sub.Items.Data[0].ID

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(subscriptionItemID),
				Price: stripe.String(newPriceID),
			},
		},
	}
	result, err := subscription.Update(subscriptionID, params)
	return result, err
}

func (c *Client) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	stripe.Key = c.config.StripeAPIKey

	sub, err := subscription.Cancel(subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
