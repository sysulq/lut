package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

//go:embed luts/*
var luts embed.FS

func init() {

	files := lo.Must(luts.ReadDir("luts"))

	lo.ForEach(files, func(item fs.DirEntry, index int) {
		if item.IsDir() {
			return
		}
		createIfNotExist(os.TempDir() + item.Name())
	})
}

func createIfNotExist(name string) {
	if file, err := os.ReadFile(name); os.IsNotExist(err) || len(file) == 0 {
		lo.Must0(os.WriteFile(name, lo.Must(luts.ReadFile(filepath.Join("luts", filepath.Base(name)))), 0644))
	}
}
