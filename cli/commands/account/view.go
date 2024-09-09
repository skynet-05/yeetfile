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
	SetEmail
	ChangePassword
	SetPasswordHint
	ManageSubscription
	PurchaseSubscription
	DeleteAccount
	Exit
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

func showPasswordHintView() {
	var changed bool
	var err error
	var passwordHint string

	passwordHintForm := func(prevErr error) (bool, error) {
		var submitted bool
		var errMsg string
		if prevErr != nil {
			errMsg = styles.ErrStyle.Render(prevErr.Error())
		}

		err := huh.NewForm(huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle("Set Password Hint")).
				Description("Set a new password hint below, or leave blank to disable."),
			huh.NewText().
				Title("Password Hint").
				Placeholder("Hint...").
				Value(&passwordHint),
			huh.NewConfirm().
				Affirmative("Submit").
				Negative("Cancel").
				Description(errMsg).
				Value(&submitted)),
		).WithTheme(styles.Theme).Run()

		if err == huh.ErrUserAborted || !submitted {
			return false, nil
		} else if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}

	changed, err = passwordHintForm(nil)
	for err != nil || changed {
		if !changed {
			break
		}

		err = changePasswordHint(passwordHint)
		if err != nil {
			changed, err = passwordHintForm(err)
		} else {
			break
		}
	}

	if changed && err == nil {
		_ = huh.NewForm(huh.NewGroup(
			utils.CreateHeader(
				"Set Password Hint",
				"Your password hint has been modified"),
			huh.NewConfirm().Affirmative("OK").Negative(""),
		)).WithTheme(styles.Theme).Run()
	}

	ShowAccountModel()
}

func showChangeEmailWarning() {
	var confirmed bool
	err := huh.NewForm(huh.NewGroup(
		utils.CreateHeader("Change Email", "An email will "+
			"be sent to your current email to initiate "+
			"the process of changing your email. Do you "+
			"want to continue?"),
		huh.NewConfirm().
			Affirmative("Continue").
			Negative("Cancel").
			Value(&confirmed),
	)).WithTheme(styles.Theme).Run()

	if err == huh.ErrUserAborted || !confirmed {
		ShowAccountModel()
	} else {
		showChangeEmailView()
	}
}

func showChangeEmailView() {
	var (
		changeID   string
		identifier string
		newEmail   string
		password   string
		confirmed  bool
	)

	response, err := globals.API.StartChangeEmail()
	if err != nil {
		utils.ShowErrorForm("Error starting email change process")
		ShowAccountModel()
		return
	}

	var idField huh.Field
	var identifierField huh.Field
	var warningNote huh.Field
	if len(response.ChangeID) > 0 {
		changeID = response.ChangeID
		idField = huh.NewNote().Title("Change ID").Description(changeID)
		identifierField = huh.NewInput().Title("Account ID").Value(&identifier)
		warningNote = huh.NewNote().
			Title("NOTE").
			Description("Once set, you will need to log in using " +
				"your email instead of your account ID.")
	} else {
		idField = huh.NewInput().
			Title("Change ID").
			Description("Paste the code from the email below:").
			Value(&changeID).Validate(
			func(s string) error {
				if len(s) == 0 {
					return errors.New("change ID cannot be empty")
				}

				return nil
			})
		identifierField = huh.NewInput().Title("Current Email").Value(&identifier)
		warningNote = huh.NewNote().
			Title("NOTE").
			Description("A confirmation email will be sent to the " +
				"new address")
	}

	showChangeEmailForm := func(errMsg string) (bool, error) {
		err = huh.NewForm(huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle("New Email Address")),
			idField,
			identifierField,
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewInput().Title("New Email").Value(&newEmail),
			warningNote,
			huh.NewConfirm().
				Affirmative("Submit").
				Negative("Cancel").
				Description(errMsg).
				Value(&confirmed),
		)).WithTheme(styles.Theme).Run()

		if err == huh.ErrUserAborted {
			return false, nil
		}

		_ = spinner.New().Title("Submitting...").Action(func() {
			err = changeEmail(identifier, password, newEmail, changeID)
		}).Run()

		return confirmed, err
	}

	submitted, err := showChangeEmailForm("")
	for err != nil {
		submitted, err = showChangeEmailForm(
			styles.ErrStyle.Render("Error: " + err.Error()))
	}

	if submitted {
		showVerifyEmailView(newEmail, "")
	} else {
		ShowAccountModel()
	}
}

func showVerifyEmailView(newEmail, errMsg string) {
	var verificationCode string

	err := huh.NewForm(huh.NewGroup(
		huh.NewNote().
			Title(utils.GenerateTitle("Verify Email")).
			Description("Enter the verification code sent to "+newEmail),
		huh.NewInput().
			Title("Verification Code").
			Value(&verificationCode).Validate(
			func(s string) error {
				if len(s) != constants.VerificationCodeLength {
					return errors.New("invalid code length")
				}

				return nil
			}),
		huh.NewConfirm().
			Description(errMsg).
			Affirmative("Submit").
			Negative(""),
	)).WithTheme(styles.Theme).Run()

	if err == huh.ErrUserAborted {
		ShowAccountModel()
		return
	}

	err = globals.API.VerifyEmail(newEmail, verificationCode)
	if err != nil {
		showVerifyEmailView(newEmail, styles.ErrStyle.Render("Error: "+err.Error()))
		return
	}

	_ = huh.NewNote().
		Title(utils.GenerateTitle("Change Email")).
		Description("Your email has been successfully changed!").Run()

	ShowAccountModel()
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
	emailLabel := "Change Email"
	emailAction := ChangeEmail
	if len(account.Email) == 0 {
		emailLabel = "Set Email"
		emailAction = SetEmail
	}

	options := []huh.Option[Action]{
		huh.NewOption(emailLabel, emailAction),
		huh.NewOption("Change Password", ChangePassword),
	}

	if len(account.Email) > 0 {
		var passwordHintOption huh.Option[Action]
		if account.HasPasswordHint {
			passwordHintOption = huh.NewOption(
				"Update / Remove Password Hint",
				SetPasswordHint)
		} else {
			passwordHintOption = huh.NewOption(
				"Set Password Hint",
				SetPasswordHint)
		}

		options = append(options, passwordHintOption)
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
	options = append(options, huh.NewOption("Exit", Exit))
	return options
}

func exitView() {}

func init() {
	actionMap = map[Action]func(){
		SetEmail:        showChangeEmailView,
		ChangeEmail:     showChangeEmailWarning,
		ChangePassword:  showChangePasswordView,
		SetPasswordHint: showPasswordHintView,
		DeleteAccount:   showAccountDeletionView,
		Exit:            exitView,
	}
}
