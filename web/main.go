package main

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"yeetfile/web/db"
	"yeetfile/web/server"
	"yeetfile/web/utils"
)

func main() {
	go db.CheckExpiry()
	go db.CheckMemberships()

	host := utils.GetEnvVar("YEETFILE_HOST", "0.0.0.0")
	port := utils.GetEnvVar("YEETFILE_PORT", "8090")

	addr := fmt.Sprintf("%s:%s", host, port)

	server.Run(addr)
}
