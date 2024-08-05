package commands

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"net"
	"yeetfile/cli/commands/auth"
	"yeetfile/cli/commands/auth/login"
	"yeetfile/cli/commands/auth/logout"
	"yeetfile/cli/commands/auth/signup"
	"yeetfile/cli/commands/download"
	"yeetfile/cli/commands/send"
	"yeetfile/cli/commands/vault"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
)

type Command string

const (
	Auth     Command = "auth"
	Signup   Command = "signup"
	Login    Command = "login"
	Logout   Command = "logout"
	Vault    Command = "vault"
	Send     Command = "send"
	Download Command = "download"
	Help     Command = "help"
)

var CommandMap = map[Command][]func(){
	Auth:     {auth.ShowAuthModel},
	Signup:   {signup.ShowSignupModel, login.ShowLoginModel},
	Login:    {login.ShowLoginModel},
	Logout:   {logout.ShowLogoutModel},
	Vault:    {vault.ShowVaultModel},
	Send:     {send.ShowSendModel},
	Download: {download.ShowDownloadModel},
	Help:     {printHelp},
}

var AuthHelp = []string{
	fmt.Sprintf("%s | Create a new YeetFile account", Signup),
	fmt.Sprintf("%s  | Log into your YeetFile account", Login),
	fmt.Sprintf("%s | Log out of your YeetFile account", Logout),
}

var ActionHelp = []string{
	fmt.Sprintf("%s    | Manage files and folders in your YeetFile Vault\n"+
		"             - Example: yeetfile vault", Vault),
	fmt.Sprintf("%s     | Create an end-to-end encrypted shareable link to a file or text\n"+
		"             - Example: yeetfile send\n"+
		"             - Example: yeetfile send path/to/file.png\n"+
		"             - Example: yeetfile send 'top secret text'", Send),
	fmt.Sprintf("%s | Download a file or text uploaded via YeetFile Send\n"+
		"             - Example: yeetfile download\n"+
		"             - Example: yeetfile download https://yeetfile.com/file_abc#top.secret.hash8\n"+
		"             - Example: yeetfile download file_abc#top.secret.hash8", Download),
}

var HelpMsg = `
Usage: yeetfile <command> [args]
`

var CommandHelpStr = `
  %s`

func printHelp() {
	HelpMsg += `
Auth Commands:`
	for _, msg := range AuthHelp {
		HelpMsg += fmt.Sprintf(CommandHelpStr, msg)
	}

	HelpMsg += `

Action Commands:`
	for _, msg := range ActionHelp {
		HelpMsg += fmt.Sprintf(CommandHelpStr, msg)
	}

	fmt.Println(HelpMsg)
	fmt.Println()
}

// Entrypoint is the main entrypoint to the CLI
func Entrypoint(args []string) {
	var command Command
	if len(args) < 2 {
		if loggedIn, err := auth.IsUserAuthenticated(); !loggedIn || err != nil {
			command = Auth
		} else if len(config.UserConfig.DefaultView) > 0 {
			command = Command(config.UserConfig.DefaultView)
		} else {
			styles.PrintErrStr("-- Missing command")
			printHelp()
			return
		}
	} else {
		command = Command(args[1])
	}

	viewFunctions, ok := CommandMap[command]
	if !ok {
		styles.PrintErrStr(fmt.Sprintf("-- Invalid command '%s'", command))
		printHelp()
		return
	}

	// Check session state and ensure server is reachable
	authErr := validateAuth()
	if _, ok := authErr.(*net.OpError); ok {
		utils.HandleCLIError("Unable to connect to the server", authErr)
		return
	} else if !isAuthCommand(command) && authErr != nil {
		styles.PrintErrStr("You are not logged in. " +
			"Use the 'login' or 'signup' commands to continue.")
		return
	}

	sessionErr := validateCurrentSession()
	if sessionErr != nil {
		errStr := fmt.Sprintf("Error validating session: %v", sessionErr)
		styles.PrintErrStr(errStr)
		return
	}

	// Set up logging output (can't log to stdout while bubbletea is running)
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	// Run view function(s)
	for _, viewFunction := range viewFunctions {
		viewFunction()
	}
}

func validateAuth() error {
	if loggedIn, err := auth.IsUserAuthenticated(); !loggedIn || err != nil {
		if err != nil {
			return err
		}
		return errors.New("not logged in")
	}

	return nil
}

func validateCurrentSession() error {
	encPrivateKey, publicKey, err := config.UserConfigPaths.GetKeys()
	if err != nil {
		errMsg := fmt.Sprintf("Error reading key files: %v\n", err)
		styles.PrintErrStr(errMsg)
	}

	cliKey := crypto.ReadCLIKey()
	if cliKey == nil || len(cliKey) == 0 {
		errMsg := fmt.Sprintf(`Missing '%[1]s' environment variable.
You must include the value returned for this variable either in your shell
config file (.bashrc, .zshrc, etc) or prefix commands with the variable (i.e.
%[1]s=xxxx yeetfile vault`, crypto.CLIKeyEnvVar)
		return errors.New(errMsg)
	}

	if privateKey, err := crypto.DecryptChunk(cliKey, encPrivateKey); err == nil {
		crypto.IngestKeys(privateKey, publicKey)
	} else {
		var cliKeyFunc func(errMsgs ...string)
		cliKeyFunc = func(errMsgs ...string) {
			cliPassword := vault.ShowVaultPasswordPromptModel(errMsgs...)
			key := crypto.DerivePBKDFKey(cliPassword, cliKey)
			if privateKey, err := crypto.DecryptChunk(key, encPrivateKey); err == nil {
				crypto.IngestKeys(privateKey, publicKey)
			} else {
				cliKeyFunc("Incorrect password")
			}
		}

		cliKeyFunc()
	}

	return nil
}

// isAuthCommand checks if the provided command is related to authentication
func isAuthCommand(cmd Command) bool {
	return cmd == Login || cmd == Signup || cmd == Logout
}
