package main

import (
	"github.com/0x7fffffff/verbatim/relay"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	relay.Start()
}
