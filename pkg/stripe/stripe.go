package stripe

import (
	"errors"

	"github.com/stripe/stripe-go/v79"
	billingsession "github.com/stripe/stripe-go/v79/billingportal/session"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/subscription"
	"github.com/stripe/stripe-go/v79/customer"
)

type Client struct {
	StripeAPIKey string
}

func NewClient(stripeAPIKey string) *Client {
	return &Client{StripeAPIKey: stripeAPIKey}
}

func (c *Client) CreateCheckoutSession(priceID string, customerAccountId string, redirectUrl string) (*stripe.CheckoutSession, error) {
	stripe.Key = c.StripeAPIKey
	checkoutParams := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(redirectUrl),
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

func (c *Client) GetSession(sessionID string) (*stripe.CheckoutSession, error) {
	stripe.Key = c.StripeAPIKey
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
	stripe.Key = c.StripeAPIKey

	params := &stripe.SubscriptionListParams{Customer: stripe.String(customerID)}
	result := subscription.List(params)

	return result, nil
}

func (c *Client) UpdateCustomerSubscription(subscriptionID string, newPriceID string) (*stripe.Subscription, error) {
	stripe.Key = c.StripeAPIKey

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
	stripe.Key = c.StripeAPIKey

	sub, err := subscription.Cancel(subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (c *Client) GetBillingPortalURL(customerID string, returnURL string) (*stripe.BillingPortalSession, error) {
	stripe.Key = c.StripeAPIKey

	portalParams := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	return billingsession.New(portalParams)
}

func (c *Client) CreateCustomer(email string, name string) (*stripe.Customer, error) {
	stripe.Key = c.StripeAPIKey

	customerParams := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	return customer.New(customerParams)
}

func (c *Client) DeleteCustomer(customerID string) error {
	stripe.Key = c.StripeAPIKey

	_, err := customer.Del(customerID, nil)
	return err
}