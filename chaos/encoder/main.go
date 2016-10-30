package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
)

func main() {
	defaults := flag.Bool("d", false, "Use default info")
	portNum := flag.Int("port", 9642, "The port on which to listen")
	username := flag.String("username", "mcs", "The user name to be required")
	password := flag.String("password", "abc123", "The password to be required")
	flag.Parse()
	if (portNum != nil && username != nil && password != nil) || (defaults != nil && *defaults) {
		fmt.Println("Serving on port:", *portNum)
		fmt.Println(*portNum, *username, *password, *defaults)
		serve(*portNum, *username, *password)
	} else {
		flag.PrintDefaults()
	}
}

var connCount = 0
var connMux = &sync.Mutex{}

func serve(port int, username, password string) {
	var ln, err = net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		fmt.Println("Unable to bind to port", port, "!")
		return
	}

	for {
		conn, err := ln.Accept()
		if err == nil {
			connMux.Lock()
			fmt.Println(connCount)
			if connCount != 0 {
				conn.Close()
			} else {
				connCount = 1
				go workConn(conn, username, password)
			}
			connMux.Unlock()
		}
		// If we had an error, just try to listen again.
	}
}

func disconnect(conn net.Conn) {
	connMux.Lock()
	conn.Close()
	connCount = 0
	connMux.Unlock()
}

func workConn(conn net.Conn, username, password string) {
	var reader = bufio.NewReader(conn)
	userInput, err := reader.ReadString('\n')
	passwordInput, err := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)
	passwordInput = strings.TrimSpace(passwordInput)
	if userInput != username || password != passwordInput || err != nil {
		disconnect(conn)
		return
	}
	fmt.Println("Connected, dawg!")
	buf := make([]byte, 512)
	for {
		if n, err := conn.Read(buf); err != nil {
			fmt.Println("Disconnected!")
			disconnect(conn)
			return
		} else {
			fmt.Print(string(buf[0:n]))
		}
	}
}
