package main

import (
	"yeetfile/utils"
	"yeetfile/web/server"
)

func main() {
	//server.TestUpload()
	//server.TestDownload()
	//fmt.Println(utils.GenFilePath())
	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port)
}
