package main

import (
	"fmt"
	"log"
	"os"
	"yeetfile/shared"
)

const outputPath = "web/static/js"
const jsConstsFile = "constants.js"
const jsEndpointsFile = "endpoints.js"

func write(path, contents string) {
	if err := os.WriteFile(path, []byte(contents), 0666); err != nil {
		log.Fatal(err)
	}
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	consts, endpoints := shared.GenerateSharedJS()
	fullPath := fmt.Sprintf("%s/%s", cwd, outputPath)
	if _, err := os.Stat(fullPath); err != nil {
		log.Fatal(err)
	}

	constsOut := fmt.Sprintf("%s/%s", fullPath, jsConstsFile)
	endpointsOut := fmt.Sprintf("%s/%s", fullPath, jsEndpointsFile)

	write(constsOut, consts)
	write(endpointsOut, endpoints)

	fmt.Printf("JS constants written to: %s\n", constsOut)
	fmt.Printf("JS endpoints written to: %s\n", endpointsOut)

}
