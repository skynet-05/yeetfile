package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"yeetfile/cli/config"
	"yeetfile/cli/utils"
)

var userConfig config.Config
var configPaths config.Paths

type Command string

const (
	Signup   Command = "signup"
	Upload   Command = "upload"
	Download Command = "download"
)

var CommandMap = map[Command]func(string){
	Signup:   signup,
	Upload:   upload,
	Download: download,
}

var HelpMap = map[Command]string{
	Signup:   signupHelp,
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
	flag.Parse()

	// Check if the user is requesting help generally or for a specific cmd
	var help bool
	flag.BoolVar(&help, "help", false, "View help")
	flag.BoolVar(&help, "h", false, "View help")

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

func upload(arg string) {
	var downloads int
	flag.IntVar(&downloads, "d", -1, "")
	flag.IntVar(&downloads, "downloads", -1, "")

	var expiration string
	flag.StringVar(&expiration, "e", "", "")
	flag.StringVar(&expiration, "expiration", "", "")

	if _, err := os.Stat(arg); err == nil {
		// Arg is a file that we should upload
		if len(expiration) == 0 {
			fmt.Println("Missing expiration argument ('-e'), " +
				"see '-h' for help with uploading.")
			return
		}

		UploadFile(arg, downloads, expiration)
	} else {
		fmt.Printf("Unable to open file: '%s'", arg)
	}
}

func download(arg string) {
	// Arg is a URL or tag for a file
	path, pepper, err := utils.ParseDownloadString(arg)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	StartDownload(path, pepper)
}

func init() {
	// Setup config dir
	var err error
	configPaths, err = config.SetupConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	userConfig, err = config.ReadConfig(configPaths)
}
