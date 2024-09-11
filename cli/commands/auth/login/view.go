package login

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"strings"
	"yeetfile/cli/api"
	"yeetfile/cli/crypto"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var randomVaultPwLabel = "Use randomly generated vault key"
var userVaultPwLabel = "Set your own vault key"
var randomVaultPwOpt = 0
var userVaultPwOpt = 1

var vaultPasswordDesc = `This will set your vault password for this session.
If you lose or forget this password, you can log out and log back in to
generate a new one.`

var cliKeyMessage = `Your CLI session key:

%s

This key is generated every time you log in, and 
WILL NOT be shown again until your next login.

You must set this variable in your environment,
and can do so in a few ways:`

var cliKeyFormat = "%s=\"%s\""

var configCLIKeyTitle = "Set in your shell's config file"
var configCLIKeyMsg = "echo '%s' >> .bashrc"

var sessionCLIKeyTitle = "Export for your current shell session"
var sessionCLIKeyMsg = "export %s"

var prefixCLIKeyTitle = "Prefix all commands with the env var"
var prefixCLIKeyMsg = "Example: %s yeetfile vault"

func ShowLoginModel() {
	var identifier string
	var password string
	var option int

	var runFunc func(errorMessages ...string) error
	runFunc = func(errMsgs ...string) error {
		loginSelected := true

		title := huh.NewNote().Title(utils.GenerateTitle("Login"))
		if len(errMsgs) > 0 {
			title.Description(styles.ErrStyle.Render(errMsgs[0]))
		}

		options := []huh.Option[int]{
			huh.NewOption(randomVaultPwLabel, randomVaultPwOpt),
			huh.NewOption(userVaultPwLabel, userVaultPwOpt),
		}

		err := huh.NewForm(
			huh.NewGroup(
				title,
				huh.NewInput().Title("Identifier").
					Description("Email or Account ID").
					Value(&identifier),
				huh.NewInput().Title("Password").
					EchoMode(huh.EchoModePassword).
					Value(&password),
				huh.NewSelect[int]().Options(options...).
					Value(&option),
				huh.NewConfirm().
					Affirmative("Log In").
					Negative("Forgot Password").
					Value(&loginSelected),
			),
		).WithTheme(styles.Theme).WithShowHelp(true).Run()

		if err != nil {
			return err
		}

		if !loginSelected {
			// User selected "forgot password"
			email := ""
			if strings.Contains(identifier, "@") {
				email = identifier
			}
			err = showForgotPasswordModel(email, "")
			if err == nil {
				return runFunc("")
			}
		}

		sessionKey, err := crypto.GenerateCLISessionKey()

		// Vault key is by default the same as the session key, unless
		// the user provides a unique vault password
		vaultKey := sessionKey
		if err != nil {
			return err
		}

		if option == userVaultPwOpt {
			vaultKeyPassword := promptVaultPassword()
			vaultKey = crypto.DerivePBKDFKey(
				[]byte(vaultKeyPassword),
				sessionKey,
			)
		}

		err = LogIn(identifier, password, "", sessionKey, vaultKey)
		if err != nil && err != api.TwoFactorError {
			return runFunc(err.Error())
		} else if err == api.TwoFactorError {
			for err == api.TwoFactorError {
				code := showTwoFactorPrompt()
				err = LogIn(identifier, password, code, sessionKey, vaultKey)
			}

			if err != nil {
				return err
			}
		}

		showCLISessionNote(string(sessionKey))
		return nil
	}

	err := runFunc()
	if err != nil && err != huh.ErrUserAborted {
		panic(err)
	}
}

func showTwoFactorPrompt() string {
	var code string
	_ = huh.NewForm(huh.NewGroup(
		utils.CreateHeader(
			"Two-Factor Enabled",
			"Enter your 2FA or recovery code below"),
		huh.NewInput().Title("2FA Code").Value(&code),
		huh.NewConfirm().Affirmative("Submit").Negative(""),
	)).WithTheme(styles.Theme).Run()

	return code
}

func showCLISessionNote(sessionKey string) {
	formattedVar := fmt.Sprintf(
		cliKeyFormat,
		shared.EscapeString(crypto.CLIKeyEnvVar),
		sessionKey)
	desc := fmt.Sprintf(cliKeyMessage, formattedVar)
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(
				utils.GenerateTitle("Vault Session Key")).
				Description(desc),
			huh.NewNote().
				Title("Option 1: "+configCLIKeyTitle).
				Description(fmt.Sprintf(configCLIKeyMsg, formattedVar)),
			huh.NewNote().
				Title("Option 2: "+sessionCLIKeyTitle).
				Description(fmt.Sprintf(sessionCLIKeyMsg, formattedVar)),
			huh.NewNote().
				Title("Option 3: "+prefixCLIKeyTitle).
				Description(fmt.Sprintf(prefixCLIKeyMsg, formattedVar)),
			huh.NewConfirm().Affirmative("OK").Negative(""))).
		WithTheme(styles.Theme).Run()
	utils.HandleCLIError("error showing session note", err)
}

func showForgotPasswordModel(email string, errMsg string) error {
	var submitted bool
	var desc string
	if len(errMsg) > 0 {
		desc = styles.ErrStyle.Render(errMsg)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle("Forgot Password")).
				Description("Enter the email address associated "+
					"with your YeetFile account below.\n\n"+
					"If you've set a password hint, it will "+
					"be emailed to you."),
			huh.NewInput().Title("Email Address").Value(&email),
			huh.NewConfirm().
				Affirmative("Submit").
				Negative("Cancel").
				Description(desc).
				Value(&submitted),
		)).WithTheme(styles.Theme).Run()

	if err == huh.ErrUserAborted || !submitted {
		return nil
	}

	err = RequestPasswordHint(email)
	if err != nil {
		return showForgotPasswordModel(email, err.Error())
	}

	_ = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle("Forgot Password")).
				Description("Your request has been submitted.\nIf a "+
					"password hint was set for this account, you will receive "+
					"an email shortly.\nIf you don't receive an email after a "+
					"few minutes and it isn't in your spam folder,\ncontact "+
					"the host for assistance."),
			huh.NewConfirm().
				Affirmative("OK").
				Negative("")),
	).WithTheme(styles.Theme).Run()

	return nil
}

func promptVaultPassword() string {
	var password string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(utils.GenerateTitle(
				"Login > Set Vault Key Password")).
				Description(vaultPasswordDesc),
			huh.NewInput().Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewInput().Title("Confirm Password").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s != password {
						return errors.New("passwords do not match")
					}

					return nil
				}),
			huh.NewConfirm().Affirmative("Submit").Negative(""),
		),
	).WithTheme(styles.Theme).Run()

	utils.HandleCLIError("error showing vault pw prompt", err)

	return password
}
