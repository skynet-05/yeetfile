package static

import "embed"

//go:embed js/*
//go:embed css/*
var Files embed.FS
