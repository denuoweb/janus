package main

import (
	"log"

	"github.com/htmlcoin/janus/cli"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	cli.Run()
}
