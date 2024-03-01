package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"yeetfile/cli/config"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

var userConfig config.Config
var configPaths config.Paths
var session string

type Command string

const (
	Signup   Command = "signup"
	Login    Command = "login"
	Logout   Command = "logout"
	Upload   Command = "upload"
	Download Command = "download"
)

var CommandMap = map[Command]func(string){
	Signup:   signup,
	Login:    login,
	Logout:   logout,
	Upload:   upload,
	Download: download,
}

var HelpMap = map[Command]string{
	Signup:   signupHelp,
	Login:    loginHelp,
	Logout:   logoutHelp,
	Upload:   uploadHelp,
	Download: downloadHelp,
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Missing input command")
		fmt.Print(mainHelp)
		return
	}

	command := os.Args[1]
	arg := os.Args[len(os.Args)-1]

	// Check if the user is requesting help generally or for a specific cmd
	var help bool
	utils.BoolFlag(&help, "help", false, os.Args)

	if help {
		helpMsg, ok := HelpMap[Command(command)]
		if ok {
			fmt.Print(helpMsg)
			return
		}

		fmt.Print(mainHelp)
		return
	}

	fn, ok := CommandMap[Command(command)]
	if !ok {
		fmt.Printf("Invalid command '%s'\n", command)
		fmt.Print(mainHelp)
		return
	}

	fn(arg)
}

func signup(_ string) {
	CreateAccount()
}

func login(_ string) {
	LoginUser(false)
}

func logout(_ string) {
	LogoutUser()
}

func upload(arg string) {
	var downloads int
	utils.IntFlag(&downloads, "downloads", 0, os.Args)

	var expiration string
	utils.StrFlag(&expiration, "expiration", "", os.Args)

	var isPlaintext bool
	utils.BoolFlag(&isPlaintext, "is-plaintext", false, os.Args)

	if len(expiration) == 0 {
		fmt.Println("Missing expiration argument ('-e'), " +
			"see '-h' for help with uploading.")
		return
	} else if downloads < 1 {
		fmt.Println("Downloads ('-d') must be set to a number " +
			"greater than 0 and less than or equal to 10.")
		return
	}

	if _, err := os.Stat(arg); err == nil {
		// Arg is a file
		if !hasValidSession() {
			fmt.Println("-- Login required")

			// Try logging user in and then repeating the request
			if LoginUser(true) {
				upload(arg)
				return
			} else {
				fmt.Println("You need to log in before uploading a file")
				return
			}
		}

		StartFileUpload(arg, downloads, expiration)
	} else {
		// Arg is a string
		StartPlaintextUpload(arg, downloads, expiration)
	}
}

func download(arg string) {
	// Arg is a URL or tag for a file
	path, pepper, err := utils.ParseDownloadString(arg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Fetch file metadata
	metadata, err := FetchMetadata(path)
	if err != nil {
		fmt.Println("Error fetching path")
		return
	}

	// Attempt first download without a password
	download, err := PrepareDownload(metadata, []byte(""), pepper)
	if err == wrongPassword {
		pw := utils.RequestPassword()
		download, err = PrepareDownload(metadata, pw, pepper)
		if err == wrongPassword {
			fmt.Println("Incorrect password")
			return
		}
	} else if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Failed to derive key to decrypt contents")
		return
	}

	// Ensure the file is what the user expects
	isPlaintext := strings.HasPrefix(path, shared.PlaintextIDPrefix)
	if !isPlaintext && download.VerifyDownload() {
		// Begin download
		err = download.DownloadFile()
		if err != nil {
			fmt.Printf("Failed to download file: %v\n", err)
		}
	} else if isPlaintext {
		err = download.DownloadPlaintext()
		if err != nil {
			fmt.Printf("Failed to download plaintext: %v\n", err)
		}
	}
}

func init() {
	// Setup config dir
	var err error
	configPaths, err = config.SetupConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	userConfig, err = config.ReadConfig(configPaths)
	session, err = config.ReadSession(configPaths)
}
