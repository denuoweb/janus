package main

import (
	"log"

	"github.com/denuoweb/janus/cli"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	cli.Run()
}
