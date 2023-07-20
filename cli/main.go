package main

import (
	"flag"
	"fmt"
	"os"
)

const domain string = "http://localhost:8090"

const helpMsg = `yeetfile

* Upload
    Args:
        -d : (Optional) Set # of times a file can be downloaded
        -e : [Required] Set the lifetime of the file, using the format
           <value><unit>, where value is a numeric and unit is one of
           the following:
               s (seconds)
               m (minutes)
               h (hours)
               d (days)
           Example:
               2d == 2 days
               3h == 3 hours
               20m == 20 minutes
    Examples:
        yeetfile -e 10d documents.zip
        yeetfile -d 10 -e 2h game.exe

* Download
    Args: None
    Examples:
        yeetfile https://yeetfile.com/d/unique.file.path
        yeetfile other.unique.path

`

func main() {
	arg := os.Args[len(os.Args)-1]

	downloads := flag.Int(
		"d",
		-1,
		"(Optional) Set # of times a file can be downloaded")
	expiration := flag.String(
		"e",
		"",
		"Set when the file expires "+
			"(default: '1d', see -h for details)")
	help := flag.Bool("h", false, "View help message")
	flag.Parse()

	if *help {
		fmt.Print(helpMsg)
		return
	} else if len(*expiration) == 0 {
		fmt.Println("Missing expiration argument ('-e'), see '-h' " +
			"for help uploading.")
		return
	}

	if arg == "config" {
		// TODO: Implement configurable backend
	} else if _, err := os.Stat(arg); err == nil {
		// Arg is a file that we should upload
		UploadFile(arg, *downloads, *expiration)
	} else {
		// Arg is (probably) a URL for a file
		StartDownload(arg)
	}
}
