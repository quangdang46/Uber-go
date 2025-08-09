package stripe

import (
	"context"
	"ride-sharing/services/payment-service/internal/domain"
	"ride-sharing/services/payment-service/pkg/types"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type stripClient struct {
	config *types.PaymentConfig
}

func NewStripeClient(config *types.PaymentConfig) domain.PaymentProcessor {

	stripe.Key = config.StripeSecretKey

	return &stripClient{
		config: config,
	}

}

func (c *stripClient) CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(c.config.SuccessURL),
		CancelURL:  stripe.String(c.config.CancelURL),
		Metadata:   (metadata),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					UnitAmount: stripe.Int64(amount),
					Currency:   stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Ride Payment"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}
	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}
