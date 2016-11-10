package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"strings"
	"sync"
)

func main() {
	dbLocation := flag.String("db", "database.db", "The location of the database file used to spin up encoders")
	flag.Parse()
	db, err := sql.Open("sqlite3", *dbLocation)
	if err != nil {
		log.Fatal(err)
		return
	}
	rows, err := db.Query(`
		Select 
			port, 
			handle, 
			password, 
			network.name 
		from encoder 
		inner join network 
			on encoder.network_id = network.id 
		where 
			ip_address = '127.0.0.1' 
			or ip_address = '::1' 
			or ip_address = 'localhost';`)
	if err != nil {
		log.Fatal(err)
		return
	}
	var wg = sync.WaitGroup{}
	for rows.Next() {
		var port int
		var handle string
		var password string
		var network_name string
		err := rows.Scan(&port, &handle, &password, &network_name)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go serve(port, handle, password, network_name, wg)
	}
	wg.Wait()
}

func serve(port int, username, password, networkName string, wg sync.WaitGroup) {
	defer wg.Done()
	var connCount = 0
	var connMux = &sync.Mutex{}
	disconnect := func(conn net.Conn) {
		connMux.Lock()
		conn.Close()
		connCount = 0
		connMux.Unlock()
	}
	var ln, err = net.Listen("tcp", fmt.Sprint(":", port))
	log.Println("Listening for", networkName, "on", port)
	if err != nil {
		fmt.Println("Unable to bind to port", port, "!")
		return
	}

	for {
		conn, err := ln.Accept()
		if err == nil {
			connMux.Lock()
			// fmt.Println(connCount)
			if connCount != 0 {
				conn.Close()
			} else {
				connCount = 1
				go workConn(conn, username, password, networkName, disconnect)
			}
			connMux.Unlock()
		}
		// If we had an error, just try to listen again.
	}
}

func workConn(conn net.Conn, username, password, networkName string, disconnect func(conn net.Conn)) {
	var reader = bufio.NewReader(conn)
	userInput, err := reader.ReadString('\n')
	passwordInput, err := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)
	passwordInput = strings.TrimSpace(passwordInput)
	if userInput != username || password != passwordInput || err != nil {
		disconnect(conn)
		return
	}
	fmt.Println("Connected for ", networkName)
	buf := make([]byte, 512)
	for {
		if n, err := conn.Read(buf); err != nil {
			fmt.Println("Disconnected for", networkName)
			disconnect(conn)
			return
		} else {
			fmt.Print(networkName+":", string(buf[0:n]))
		}
	}
}
