package main

import (
	"github.com/0x7fffffff/verbatim/dashboard"
	"github.com/0x7fffffff/verbatim/relay"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	go relay.Start()
	dashboard.Start()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
