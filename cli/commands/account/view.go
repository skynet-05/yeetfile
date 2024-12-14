package account

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh/spinner"
	"github.com/mdp/qrterminal/v3"
	"strconv"
	"strings"
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
	SetTwoFactor
	DeleteTwoFactor
	PurchaseUpgrade
	RecyclePaymentID
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

func showSetTwoFactorView() {
	var newTOTP shared.NewTOTP
	var err error
	_ = spinner.New().Title("Generating 2FA...").Action(func() {
		newTOTP, err = globals.API.Generate2FA()
	}).Run()

	if err != nil {
		utils.ShowErrorForm("Error generating 2FA")
		ShowAccountModel()
		return
	}

	builder := strings.Builder{}
	qrterminal.GenerateHalfBlock(newTOTP.URI, qrterminal.L, &builder)

	submitted := true
	var totpCode string
	var response shared.SetTOTPResponse

	var show2FASetup func(errMsg string) error
	show2FASetup = func(errMsg string) error {
		err = huh.NewForm(huh.NewGroup(
			utils.CreateHeader("Enable 2FA", ""),
			huh.NewNote().Title("Secret: "+newTOTP.Secret).Description(builder.String()),
			huh.NewInput().
				Title("2FA Code").
				Description("Use your authenticator app to scan the "+
					"QR above, then input the 6-digit code below").
				Value(&totpCode).
				Validate(func(s string) error {
					if len(s) == 6 || len(s) == 0 {
						return nil
					}

					return errors.New("code must be 6 digits")
				}),
			huh.NewConfirm().
				Affirmative("Submit").
				Negative("Cancel").
				Description(errMsg).
				Value(&submitted),
		)).WithTheme(styles.Theme).Run()

		if err != nil {
			return err
		} else if !submitted {
			return huh.ErrUserAborted
		}

		_ = spinner.New().Title("Setting up 2FA...").Action(func() {
			response, err = globals.API.Finalize2FA(shared.SetTOTP{
				Secret: newTOTP.Secret,
				Code:   totpCode,
			})
		}).Run()

		if err != nil {
			return show2FASetup(styles.ErrStyle.Render(err.Error()))
		}

		return nil
	}

	err = show2FASetup("")
	if err == huh.ErrUserAborted {
		ShowAccountModel()
		return
	}

	var recoveryCodes string
	for _, code := range response.RecoveryCodes {
		recoveryCodes += fmt.Sprintf("\n%s", code)
	}

	_ = huh.NewForm(huh.NewGroup(
		utils.CreateHeader(
			"2FA Enabled", "Two-factor authentication has been "+
				"enabled for your account!"),
		huh.NewNote().
			Title("Recovery Codes").
			Description(utils.GenerateWrappedText("These are "+
				"ONE-TIME recovery codes that "+
				"can be used to log in if you ever lose your "+
				"2FA app. Write these down somewhere safe.\n"+
				recoveryCodes)),
		huh.NewConfirm().Affirmative("OK").Negative(""),
	)).WithTheme(styles.Theme).Run()

	ShowAccountModel()
}

func showDeleteTwoFactorView() {
	var code string
	var confirmed bool

	var deleteFunc func(string) error
	deleteFunc = func(errMsg string) error {
		err := huh.NewForm(huh.NewGroup(
			utils.CreateHeader(
				"Disable 2FA", "To disable 2FA, enter your "+
					"2FA or recovery code below."),
			huh.NewInput().Title("2FA Code").Value(&code),
			huh.NewConfirm().
				Affirmative("Disable 2FA").
				Negative("Cancel").
				Description(errMsg).
				Value(&confirmed),
		)).WithTheme(styles.Theme).Run()

		if err != nil {
			return err
		} else if !confirmed {
			return huh.ErrUserAborted
		}

		_ = spinner.New().Title("Disabling 2FA...").Action(func() {
			err = globals.API.Disable2FA(code)
		}).Run()

		if err != nil {
			return deleteFunc(styles.ErrStyle.Render(err.Error()))
		}

		return nil
	}

	err := deleteFunc("")
	if err == huh.ErrUserAborted {
		ShowAccountModel()
		return
	} else if confirmed && err == nil {
		_ = huh.NewForm(huh.NewGroup(
			utils.CreateHeader(
				"2FA Disabled",
				"Your account 2FA has been disabled"),
			huh.NewConfirm().Affirmative("OK").Negative(""),
		)).WithTheme(styles.Theme).Run()
	}

	ShowAccountModel()
}

func showRecyclePaymentIDView() {
	title := utils.GenerateTitle("Recycle Payment ID")
	desc := "Recycling your payment ID is a privacy feature that removes " +
		"the ability for YeetFile to connect your account with past " +
		"payments to upgrade your account. This should only be done if " +
		"you've previously made a payment to YeetFile and want to remove " +
		"the record of this transaction."
	desc = utils.GenerateWrappedText(desc)

	var confirmed bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewNote().Title(title).Description(desc),
		huh.NewConfirm().
			Affirmative("Confirm").
			Negative("Cancel").Value(&confirmed),
	)).WithTheme(styles.Theme).Run()

	if err != nil || !confirmed {
		ShowAccountModel()
		return
	}

	_ = spinner.New().Title("Recycling payment ID...").Action(func() {
		err = globals.API.RecyclePaymentID()
	}).Run()

	if err != nil {
		utils.ShowErrorForm(err.Error())
	}

	ShowAccountModel()
}

func generateSelectOptions(
	account shared.AccountResponse,
) []huh.Option[Action] {
	var options []huh.Option[Action]

	changePasswordAction := huh.NewOption("Change Password", ChangePassword)
	if globals.ServerInfo.EmailConfigured {
		emailLabel := "Change Email"
		emailAction := ChangeEmail
		if len(account.Email) == 0 {
			emailLabel = "Set Email"
			emailAction = SetEmail
		}

		options = []huh.Option[Action]{
			huh.NewOption(emailLabel, emailAction),
			changePasswordAction,
		}
	} else {
		options = []huh.Option[Action]{changePasswordAction}
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

	var twoFactorOption huh.Option[Action]
	if account.Has2FA {
		twoFactorOption = huh.NewOption("Remove 2FA", DeleteTwoFactor)
	} else {
		twoFactorOption = huh.NewOption("Enable 2FA", SetTwoFactor)
	}

	options = append(options, twoFactorOption)

	if globals.ServerInfo.BillingEnabled && len(globals.ServerInfo.Upgrades) > 0 {
		options = append(
			options,
			huh.NewOption("Upgrade Account", PurchaseUpgrade))
	}

	options = append(options, huh.NewOption("Recycle Payment ID", RecyclePaymentID))
	options = append(options, huh.NewOption("Delete Account", DeleteAccount))
	options = append(options, huh.NewOption("Exit", Exit))
	return options
}

func showUpgradeView() {
	const (
		switchYearlyMonthly = -1
		cancel              = -2
	)

	var availableUpgrades []shared.Upgrade
	var upgradeFunc func(bool) (int, error)
	upgradeFunc = func(isYearly bool) (int, error) {
		var switchOption huh.Option[int]
		var selected int
		if isYearly {
			switchOption = huh.NewOption(
				"Switch to monthly",
				switchYearlyMonthly)
		} else {
			switchOption = huh.NewOption(
				"Switch to yearly (2 months free)",
				switchYearlyMonthly)
		}

		if isYearly {
			availableUpgrades = globals.ServerInfo.YearUpgrades
		} else {
			availableUpgrades = globals.ServerInfo.MonthUpgrades
		}

		options := []huh.Option[int]{
			huh.NewOption("Cancel", cancel),
			switchOption,
		}

		fields := []huh.Field{
			huh.NewNote().
				Title(utils.GenerateTitle("Upgrade Account")).
				Description("Note: All upgrades are one-time " +
					"purchases and do not auto-renew."),
		}

		for i, upgrade := range availableUpgrades {
			upgradeOption := huh.NewOption(upgrade.Name+" ->", i)
			upgradeDesc := utils.GenerateDescription(
				generateUpgradeDesc(upgrade), 25)
			upgradeNote := huh.NewNote().
				Title(upgrade.Name).
				Description(upgradeDesc)
			options = append(options, upgradeOption)
			fields = append(fields, upgradeNote)
		}

		fields = append(fields, huh.NewSelect[int]().Options(options...).Value(&selected))
		err := huh.NewForm(huh.NewGroup(fields...)).WithTheme(styles.Theme).Run()

		if err == nil && selected == switchYearlyMonthly {
			return upgradeFunc(!isYearly)
		}

		return selected, err
	}

	selected, err := upgradeFunc(false)
	if err == huh.ErrUserAborted || selected == cancel {
		ShowAccountModel()
	} else {
		showCheckoutModel(availableUpgrades[selected])
	}
}

func showCheckoutModel(upgrade shared.Upgrade) {
	const (
		stripe int = iota
		btcpay
		back
	)

	var (
		selected int
		duration string
	)

	quantity := "1"
	upgradeName := upgrade.Name
	if upgrade.Duration == constants.DurationYear {
		upgradeName += " (Year)"
		duration = "years"
	} else {
		upgradeName += " (Month)"
		duration = "months"
	}

	subDesc := generateUpgradeDesc(upgrade)
	subDesc += "\n\n" + utils.GenerateWrappedText(
		"Select a checkout option below. This will provide a link "+
			"to complete your purchase on the web.") + "\n\n" +
		utils.GenerateWrappedText("You do not need to log "+
			"into your YeetFile account on the web to complete the transaction.")

	options := []huh.Option[int]{
		huh.NewOption("Stripe Checkout (USD, CAD, GBP, etc)", stripe),
		huh.NewOption("BTCPay Checkout (BTC or XMR)", btcpay),
		huh.NewOption("Go Back", back),
	}

	err := huh.NewForm(huh.NewGroup(
		huh.NewNote().
			Title(utils.GenerateTitle("Upgrade")).
			Description(utils.GenerateDescriptionSection(
				upgradeName,
				subDesc,
				30)),
		huh.NewInput().
			Title("Quantity").
			Description("Number of "+duration).
			Value(&quantity).Validate(
			func(s string) error {
				intVal, err := strconv.Atoi(s)
				if err != nil {
					return err
				} else if intVal < 1 {
					return errors.New("quantity must be > 1")
				} else if intVal > 12 {
					return errors.New("quantity must be < 12")
				}

				return nil
			}),
		huh.NewSelect[int]().Options(options...).Value(&selected),
	)).WithTheme(styles.Theme).Run()

	var link string
	if err == huh.ErrUserAborted || selected == back {
		showUpgradeView()
		return
	} else if selected == stripe {
		link, err = globals.API.InitStripeCheckout(upgrade.Tag, quantity)
		if err == nil {
			showCheckoutLinkModel(link)
		}
	} else if selected == btcpay {
		link, err = globals.API.InitBTCPayCheckout(upgrade.Tag, quantity)
		if err == nil {
			showCheckoutLinkModel(link)
		}
	}

	if err != nil {
		utils.ShowErrorForm("Error generating checkout link")
		showUpgradeView()
	}
}

func showCheckoutLinkModel(link string) {
	desc := fmt.Sprintf("Use the link below to finish checkout:\n\n%s\n\n"+
		"When you are finished checking out, you can return to the CLI.",
		shared.EscapeString(link))

	_ = huh.NewForm(huh.NewGroup(
		huh.NewNote().
			Title(utils.GenerateTitle("Checkout")).
			Description(desc),
		huh.NewConfirm().Affirmative("OK").Negative("")),
	).WithTheme(styles.Theme).Run()
}

func exitView() {}

func init() {
	actionMap = map[Action]func(){
		SetEmail:         showChangeEmailView,
		ChangeEmail:      showChangeEmailWarning,
		ChangePassword:   showChangePasswordView,
		SetPasswordHint:  showPasswordHintView,
		SetTwoFactor:     showSetTwoFactorView,
		PurchaseUpgrade:  showUpgradeView,
		DeleteTwoFactor:  showDeleteTwoFactorView,
		RecyclePaymentID: showRecyclePaymentIDView,
		DeleteAccount:    showAccountDeletionView,
		Exit:             exitView,
	}
}
