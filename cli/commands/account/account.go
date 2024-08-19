package account

import (
	"errors"
	"fmt"
	"time"
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
		PrevLoginKeyHash: oldLoginKeyHash,
		NewLoginKeyHash:  newLoginKeyHash,
		ProtectedKey:     newProtectedKey,
	})
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
