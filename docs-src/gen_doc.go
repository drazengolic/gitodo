package main

import (
	"log"

	"github.com/drazengolic/gitodo/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cmd.RootCmd, ".")
	if err != nil {
		log.Fatal(err)
	}
}
