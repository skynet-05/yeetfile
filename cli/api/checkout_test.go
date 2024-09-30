//go:build server_test

package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"yeetfile/backend/server/subscriptions"
)

func TestCheckoutInit(t *testing.T) {
	info, err := UserA.context.GetServerInfo()
	assert.Nil(t, err)

	if info.BillingEnabled {
		account, err := UserA.context.GetAccountInfo()
		assert.Nil(t, err)

		link, err := UserA.context.InitStripeCheckout(subscriptions.MonthlyNovice)
		assert.Nil(t, err)
		assert.Contains(t, link, account.PaymentID)

		link, err = UserA.context.InitBTCPayCheckout(subscriptions.MonthlyRegular)
		assert.Nil(t, err)
		assert.Contains(t, link, account.PaymentID)
	}
}
