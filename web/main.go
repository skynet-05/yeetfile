package main

import (
	"yeetfile/db"
	"yeetfile/utils"
	"yeetfile/web/server"
)

func main() {
	go db.CheckExpiry()

	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port)
}
