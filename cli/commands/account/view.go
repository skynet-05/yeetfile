package account

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh/spinner"
	"time"
	"yeetfile/cli/globals"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"

	"github.com/charmbracelet/huh"
)

type Action int

const (
	ChangeEmail Action = iota
	ChangePassword
	ManageSubscription
	PurchaseSubscription
	DeleteAccount
)

var actionMap map[Action]func()

func ShowAccountModel() {
	account, accountDetails := FetchAccountDetails()
	options := generateSelectOptions(account)
	var action Action

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle("Account")).
				Description(utils.GenerateDescriptionSection(
					"Info",
					accountDetails, 21)),
			huh.NewSelect[Action]().
				Title("Actions").
				Options(options...).
				Value(&action),
		)).WithTheme(styles.Theme).Run()
	utils.HandleCLIError("Error displaying account model", err)

	actionViewFunc, ok := actionMap[action]
	if ok {
		actionViewFunc()
	}
}

func showChangePasswordView() {
	var identifier string
	var currentPassword string
	var newPassword string
	var confirmed bool

	var changed bool
	var formErr error

	changePasswordForm := func(prevErr error) (bool, error) {
		var errMsg string
		if prevErr != nil {
			errMsg = prevErr.Error()
		}
		err := huh.NewForm(huh.NewGroup(
			huh.NewNote().
				Title("Change Password").
				Description("Enter your current login"),
			huh.NewInput().
				Title("Identifier").
				Placeholder("Email / Account ID").
				Value(&identifier),
			huh.NewInput().
				Title("Current Password").
				EchoMode(huh.EchoModePassword).
				Value(&currentPassword),
			huh.NewInput().
				Title("New Password").
				EchoMode(huh.EchoModePassword).
				Value(&newPassword),
			huh.NewInput().
				Title("Confirm New Password").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == newPassword {
						return nil
					}

					return errors.New("passwords don't match")
				}),
			huh.NewConfirm().
				Description(styles.ErrStyle.Render(errMsg)).
				Affirmative("Change Password").
				Negative("Cancel").
				Value(&confirmed))).WithTheme(styles.Theme).Run()

		utils.HandleCLIError("Error showing password form", err)

		if !confirmed {
			ShowAccountModel()
			return false, nil
		}

		_ = spinner.New().Title("Changing password...").Action(func() {
			err = changePassword(identifier, currentPassword, newPassword)
		}).Run()

		return err == nil, err
	}

	changed, formErr = changePasswordForm(nil)
	for formErr != nil {
		changed, formErr = changePasswordForm(formErr)
	}

	if changed {
		err := huh.NewForm(huh.NewGroup(
			huh.NewNote().Title("Change Password").
				Description("Your password has successfully been changed."),
			huh.NewConfirm().
				Affirmative("OK").
				Negative("")),
		).WithTheme(styles.Theme).Run()
		utils.HandleCLIError("Error showing pw confirmation", err)
		ShowAccountModel()
	} else {
		ShowAccountModel()
	}
}

func showAccountDeletionView() {
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
) []huh.Option[Action] {
	options := []huh.Option[Action]{
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

func init() {
	actionMap = map[Action]func(){
		ChangePassword: showChangePasswordView,
		DeleteAccount:  showAccountDeletionView,
	}
}
