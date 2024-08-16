package account

import (
	"fmt"
	"time"
	"yeetfile/cli/globals"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

func getSubscriptionString(exp time.Time) string {
	if exp.Year() < 2024 {
		return "Inactive"
	} else if exp.Before(time.Now()) {
		return "Expired on " + exp.Format(time.DateOnly)
	} else {
		return "Active (expires " + utils.LocalTimeFromUTC(exp).
			Format(time.DateOnly) + ")"
	}
}

func getStorageString(used, available int, isSend bool) string {
	if available == 0 && used == 0 {
		return "None (requires subscription)"
	} else if available == 0 && used > 0 {
		return fmt.Sprintf("%s used", shared.ReadableFileSize(used))
	} else {
		var monthIndicator string
		if isSend {
			monthIndicator = "/ month"
		}
		return fmt.Sprintf("%s / %s %s (%s remaining)",
			shared.ReadableFileSize(used),
			shared.ReadableFileSize(available),
			monthIndicator,
			shared.ReadableFileSize(available-used))
	}
}

func FetchAccountDetails() (shared.AccountResponse, string) {
	account, err := globals.API.GetAccountInfo()
	if err != nil {
		return account,
			fmt.Sprintf("Error fetching account details: %v\n", err)
	}

	subscriptionStr := getSubscriptionString(account.SubscriptionExp)
	storageStr := getStorageString(account.StorageUsed, account.StorageAvailable, false)
	sendStr := getStorageString(account.SendUsed, account.SendAvailable, true)

	if len(account.Email) == 0 {
		account.Email = "None"
	}

	accountDetails := fmt.Sprintf(""+
		"Email: %s\n"+
		"Vault: %s\n"+
		"Send:  %s\n\n"+
		"Subscription: %s",
		account.Email,
		storageStr,
		sendStr,
		subscriptionStr)

	return account, accountDetails
}
