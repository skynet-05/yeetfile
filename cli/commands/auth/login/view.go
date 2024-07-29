package login

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"yeetfile/cli/crypto"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var useRandomVaultPassword = "Use randomly generated vault key"
var useUserVaultPassword = "Set your own vault key"

var vaultPasswordDesc = `This will set your vault password for this session.
If you lose or forget this password, you can log out and log back in to
generate a new one.`

var cliKeyMessage = `Your CLI session key:

%[2]s=%[1]s

You must set this variable in your environment,
and can do so in a few ways:

- Set in your shell's config file
  Ex: echo "%[2]s=%[1]s" >> .bashrc
OR
- Export for your current shell session
  Ex: export %[2]s=%[1]s
OR
- Prefix all commands with the env var
  Ex: %[2]s=%[1]s yeetfile vault

This key is generated every time you log in, and 
won't be shown again until your next login.
`

func ShowLoginModel() {
	var identifier string
	var password string
	var option string

	var runFunc func(errorMessages ...string) error
	runFunc = func(errMsgs ...string) error {
		title := huh.NewNote().Title(utils.GenerateTitle("Login"))
		if len(errMsgs) > 0 {
			title.Description(styles.ErrStyle.Render(errMsgs[0]))
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
				huh.NewSelect[string]().Options(
					huh.NewOptions(
						useRandomVaultPassword,
						useUserVaultPassword)...).
					Value(&option),
				huh.NewConfirm().Affirmative("Log In").Negative(""),
			),
		).WithTheme(styles.Theme).WithShowHelp(true).Run()

		if err != nil {
			return err
		}

		sessionKey, err := crypto.GenerateCLISessionKey()

		// Vault key is by default the same as the session key, unless
		// the user provides a unique vault password
		vaultKey := sessionKey
		if err != nil {
			return err
		}

		if option == useUserVaultPassword {
			vaultKeyPassword := promptVaultPassword()
			vaultKey = crypto.DerivePBKDFKey(
				[]byte(vaultKeyPassword),
				sessionKey,
			)
		}

		err = LogIn(identifier, password, vaultKey)
		if err != nil {
			return runFunc(err.Error())
		}

		showCLISessionNote(string(sessionKey))
		return nil
	}

	err := runFunc()
	if err != nil && err != huh.ErrUserAborted {
		panic(err)
	}
}

func showCLISessionNote(sessionKey string) {
	msg := fmt.Sprintf(cliKeyMessage, sessionKey, shared.EscapeString(crypto.CLIKeyEnvVar))
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(
				utils.GenerateTitle("Vault Session Key")).
				Description(msg),
			huh.NewConfirm().Affirmative("OK").Negative(""))).Run()
	utils.HandleCLIError("error showing session note", err)
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
