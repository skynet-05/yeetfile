package main

import (
	"yeetfile/utils"
	"yeetfile/web/server"
)

func main() {
	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port)
}
