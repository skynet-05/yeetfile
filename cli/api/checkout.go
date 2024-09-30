package api

import (
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/requests"
	"yeetfile/cli/utils"
	"yeetfile/shared/endpoints"
)

// InitStripeCheckout produces a link that the user can use to check out via Stripe
func (ctx *Context) InitStripeCheckout(subType string) (string, error) {
	url := fmt.Sprintf("%s?type=%s", endpoints.StripeCheckout.Format(ctx.Server), subType)
	resp, err := requests.GetRequest(ctx.Session, url)
	if err != nil {
		return "", err
	} else if resp.StatusCode > http.StatusBadRequest {
		return "", utils.ParseHTTPError(resp)
	}

	redirect := resp.Header.Get("Location")
	if len(redirect) == 0 {
		return "", errors.New("missing checkout link in response")
	}

	return redirect, nil
}

// InitBTCPayCheckout produces a link that the user can use to check out via BTCPay
func (ctx *Context) InitBTCPayCheckout(subType string) (string, error) {
	url := fmt.Sprintf("%s?type=%s", endpoints.BTCPayCheckout.Format(ctx.Server), subType)
	resp, err := requests.GetRequest(ctx.Session, url)
	if err != nil {
		return "", err
	} else if resp.StatusCode > http.StatusBadRequest {
		return "", utils.ParseHTTPError(resp)
	}

	redirect := resp.Header.Get("Location")
	if len(redirect) == 0 {
		return "", errors.New("missing checkout link in response")
	}

	return redirect, nil
}
