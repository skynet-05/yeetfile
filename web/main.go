package main

import (
	"embed"
	"yeetfile/web/db"
	"yeetfile/web/server"
	"yeetfile/web/utils"
)

//go:embed static/js/*
//go:embed static/css/*
var staticFiles embed.FS

func main() {
	go db.CheckExpiry()

	port := utils.GetEnvVar("YEETFILE_PORT", "8090")
	server.Run(port, staticFiles)
}
