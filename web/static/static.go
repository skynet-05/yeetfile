package static

import "embed"

//go:embed js/*
//go:embed css/*
//go:embed img/*
var StaticFiles embed.FS

//go:embed stream_saver/sw.js
//go:embed stream_saver/StreamSaver.js
//go:embed stream_saver/mitm.html
var StreamSaverFiles embed.FS
