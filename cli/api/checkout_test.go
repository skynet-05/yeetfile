//go:build server_test

package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckoutInit(t *testing.T) {
	info, err := UserA.context.GetServerInfo()
	assert.Nil(t, err)

	if info.BillingEnabled && len(info.Upgrades) > 0 {
		account, err := UserA.context.GetAccountInfo()
		assert.Nil(t, err)

		if info.StripeEnabled {
			link, err := UserA.context.InitStripeCheckout(
				info.Upgrades[0].Tag, "1")
			assert.Nil(t, err)
			assert.Contains(t, link, account.PaymentID)
		}

		if info.BTCPayEnabled {
			link, err := UserA.context.InitBTCPayCheckout(
				info.Upgrades[0].Tag, "1")
			assert.Nil(t, err)
			assert.Contains(t, link, account.PaymentID)
		}
	}
}
