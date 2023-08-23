package main

import (
	"embed"
	"yeetfile/db"
	"yeetfile/utils"
	"yeetfile/web/server"
)

//go:embed static/js/*
//go:embed static/css/*
var staticFiles embed.FS

func main() {
	go db.CheckExpiry()

	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port, staticFiles)
}
