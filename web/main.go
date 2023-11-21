package main

import (
	_ "github.com/joho/godotenv/autoload"
	"yeetfile/web/db"
	"yeetfile/web/server"
	"yeetfile/web/utils"
)

func main() {
	go db.CheckExpiry()

	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port)
}
