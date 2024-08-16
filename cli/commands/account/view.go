package account

import (
	"fmt"
	"time"
	"yeetfile/cli/globals"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"

	"github.com/charmbracelet/huh"
)

type AccountAction int

const (
	ChangeEmail AccountAction = iota
	ChangePassword
	ManageSubscription
	PurchaseSubscription
	DeleteAccount
)

func ShowAccountModel() {
	account, accountDetails := FetchAccountDetails()
	options := generateSelectOptions(account)
	var action AccountAction

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle("Account")).
				Description(utils.GenerateDescriptionSection(
					"Info",
					accountDetails, 21)),
			huh.NewSelect[AccountAction]().
				Title("Actions").
				Options(options...).
				Value(&action),
		)).WithTheme(styles.Theme).Run()
	utils.HandleCLIError("Error displaying account model", err)

	switch action {
	case DeleteAccount:
		startAccountDeletion()
		return
	}
}

func startAccountDeletion() {
	deletionFunc := func(errMsg string) (bool, string) {
		var id string
		var confirm bool
		msg := "Deleting your account cannot be undone. To confirm, \n" +
			"please enter your login email or account ID below to \n" +
			"begin deletion."

		err := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title(utils.GenerateTitle("Delete Account")).
					Description(msg),
				huh.NewInput().
					Title("Email / Account ID").
					Value(&id),
				huh.NewConfirm().
					Affirmative("Delete My Account").
					Negative("Cancel").
					Description(errMsg).
					Value(&confirm),
			)).WithTheme(styles.Theme).Run()
		utils.HandleCLIError("Error displaying account deletion", err)

		return confirm, id
	}

	confirmed, id := deletionFunc("")
	if !confirmed {
		ShowAccountModel()
		return
	}

	err := globals.API.DeleteAccount(id)
	for err != nil {
		confirmed, id = deletionFunc("Error: " + err.Error())
		if !confirmed {
			ShowAccountModel()
			return
		}

		err = globals.API.DeleteAccount(id)
	}

	_ = globals.Config.Reset()
	fmt.Println("Your YeetFile account has been deleted.")
}

func generateSelectOptions(
	account shared.AccountResponse,
) []huh.Option[AccountAction] {
	options := []huh.Option[AccountAction]{
		huh.NewOption("Change Email", ChangeEmail),
		huh.NewOption("Change Password", ChangePassword),
	}

	if account.SubscriptionExp.After(time.Now()) &&
		account.SubscriptionMethod == constants.SubMethodStripe {
		options = append(
			options,
			huh.NewOption("Manage Subscription", ManageSubscription))
	} else {
		options = append(
			options,
			huh.NewOption("Purchase Subscription", PurchaseSubscription))
	}

	options = append(options, huh.NewOption("Delete Account", DeleteAccount))
	return options
}
