package static

import (
	"bytes"
	"embed"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

//go:embed js/*
//go:embed css/*
//go:embed img/*
//go:embed icons/*.svg
var StaticFiles embed.FS

//go:embed stream_saver/sw.js
//go:embed stream_saver/StreamSaver.js
//go:embed stream_saver/mitm.html
var StreamSaverFiles embed.FS

var MinifiedFiles map[string][]byte

type minifyFn func(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error

func minifyFile(reader io.Reader, fn minifyFn) []byte {
	var buf bytes.Buffer
	if err := fn(minify.New(), &buf, reader, nil); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func minifyStaticFiles(assetType string, fn minifyFn) {
	dir, err := StaticFiles.ReadDir(assetType)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range dir {
		if !file.IsDir() && strings.HasSuffix(file.Name(), assetType) {
			originalFile, err := StaticFiles.Open(filepath.Join(assetType, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			defer func(originalFile fs.File) {
				err := originalFile.Close()
				if err != nil {

				}
			}(originalFile)

			reader := io.Reader(originalFile)

			minifiedBytes := minifyFile(reader, fn)
			MinifiedFiles[file.Name()] = minifiedBytes
		}
	}
}

func init() {
	log.Println("Minifying static assets...")
	MinifiedFiles = make(map[string][]byte)
	minifyStaticFiles("js", js.Minify)
	minifyStaticFiles("css", css.Minify)
}
