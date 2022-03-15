package main

import (
	"embed"
	"io/fs"

	"github.com/ThalesGroup/besec/cmd"
)

// Statically embed a copy of the frontend in var UIDir
//go:embed ui/build
var embedded embed.FS

func main() {
	ui, err := fs.Sub(embedded, "ui/build")
	if err != nil {
		panic("Couldn't open ui/build dir")
	}

	cmd.Execute(ui)
}
