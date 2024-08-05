package main

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"yeetfile/backend/db"
	"yeetfile/backend/server"
	"yeetfile/backend/utils"
)

func main() {
	defer db.Close()
	go db.CheckExpiry()
	go db.CheckMemberships()

	host := utils.GetEnvVar("YEETFILE_HOST", "localhost")
	port := utils.GetEnvVar("YEETFILE_PORT", "8090")

	addr := fmt.Sprintf("%s:%s", host, port)

	server.Run(addr)
}
