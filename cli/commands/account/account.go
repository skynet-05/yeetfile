package account

import (
	"errors"
	"fmt"
	"time"
	"yeetfile/backend/server/subscriptions"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

type ChangePasswordForm struct {
	Identifier  string
	OldPassword string
	NewPassword string
}

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

func getStorageString(used, available int64, isSend bool) string {
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

func changePassword(identifier, password, newPassword string) error {
	userKey := crypto.GenerateUserKey([]byte(identifier), []byte(password))
	oldLoginKeyHash := crypto.GenerateLoginKeyHash(userKey, []byte(password))

	newUserKey := crypto.GenerateUserKey([]byte(identifier), []byte(newPassword))
	newLoginKeyHash := crypto.GenerateLoginKeyHash(newUserKey, []byte(newPassword))

	protectedKey, err := globals.API.GetUserProtectedKey()
	if err != nil {
		return errors.New("error fetching protected key")
	}

	privateKey, err := crypto.DecryptChunk(userKey, protectedKey)
	if err != nil {
		return errors.New("error decrypting protected key")
	}

	newProtectedKey, err := crypto.EncryptChunk(newUserKey, privateKey)
	if err != nil {
		return errors.New("error encrypting private key")
	}

	return globals.API.ChangePassword(shared.ChangePassword{
		OldLoginKeyHash: oldLoginKeyHash,
		NewLoginKeyHash: newLoginKeyHash,
		ProtectedKey:    newProtectedKey,
	})
}

func changePasswordHint(passwordHint string) error {
	return globals.API.ChangePasswordHint(passwordHint)
}

func changeEmail(identifier, password, newEmail, changeID string) error {
	userKey := crypto.GenerateUserKey([]byte(identifier), []byte(password))
	oldLoginKeyHash := crypto.GenerateLoginKeyHash(userKey, []byte(password))

	newUserKey := crypto.GenerateUserKey([]byte(newEmail), []byte(password))
	newLoginKeyHash := crypto.GenerateLoginKeyHash(newUserKey, []byte(password))

	protectedKey, err := globals.API.GetUserProtectedKey()
	if err != nil {
		return errors.New("error fetching protected key")
	}

	privateKey, err := crypto.DecryptChunk(userKey, protectedKey)
	if err != nil {
		return errors.New("error decrypting protected key")
	}

	newProtectedKey, err := crypto.EncryptChunk(newUserKey, privateKey)
	if err != nil {
		return errors.New("error encrypting private key")
	}

	return globals.API.ChangeEmail(shared.ChangeEmail{
		NewEmail:        newEmail,
		OldLoginKeyHash: oldLoginKeyHash,
		NewLoginKeyHash: newLoginKeyHash,
		ProtectedKey:    newProtectedKey,
	}, changeID)
}

func FetchAccountDetails() (shared.AccountResponse, string) {
	account, err := globals.API.GetAccountInfo()
	if err != nil {
		msg := fmt.Sprintf("Error fetching account details: %v\n", err)
		return account, msg
	}

	subscriptionStr := getSubscriptionString(account.SubscriptionExp)
	storageStr := getStorageString(account.StorageUsed, account.StorageAvailable, false)
	sendStr := getStorageString(account.SendUsed, account.SendAvailable, true)

	emailStr := account.Email
	if len(account.Email) == 0 {
		emailStr = "None"
	}

	passwordHintStr := "Not Set"
	if account.HasPasswordHint {
		passwordHintStr = "Enabled"
	}

	twoFactorStr := "Not Set"
	if account.Has2FA {
		twoFactorStr = "Enabled"
	}

	accountDetails := fmt.Sprintf(""+
		"Email: %s\n"+
		"Vault: %s\n"+
		"Send:  %s\n\n"+
		"Subscription: %s\n"+
		"Password Hint: %s\n"+
		"Two-Factor: %s\n"+
		"Payment ID: %s",
		shared.EscapeString(emailStr),
		storageStr,
		sendStr,
		subscriptionStr,
		passwordHintStr,
		twoFactorStr,
		shared.EscapeString(account.PaymentID))

	return account, accountDetails
}

func generateSubDesc(subType string, isYearly bool) string {
	var duration string
	var price int
	if isYearly {
		duration = "year"

		switch subType {
		case subscriptions.TypeNovice:
			price = subscriptions.PriceMapping[subscriptions.YearlyNovice]
		case subscriptions.TypeRegular:
			price = subscriptions.PriceMapping[subscriptions.YearlyRegular]
		case subscriptions.TypeAdvanced:
			price = subscriptions.PriceMapping[subscriptions.YearlyAdvanced]
		}
	} else {
		duration = "month"

		switch subType {
		case subscriptions.TypeNovice:
			price = subscriptions.PriceMapping[subscriptions.MonthlyNovice]
		case subscriptions.TypeRegular:
			price = subscriptions.PriceMapping[subscriptions.MonthlyRegular]
		case subscriptions.TypeAdvanced:
			price = subscriptions.PriceMapping[subscriptions.MonthlyAdvanced]
		}
	}

	storage := shared.ReadableFileSize(subscriptions.StorageAmountMap[subType])
	send := shared.ReadableFileSize(subscriptions.SendAmountMap[subType])

	return shared.EscapeString(fmt.Sprintf(`- %s Vault Storage
- Send %s/month

 ** $%d/%s **`, storage, send, price, duration))
}
