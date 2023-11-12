package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":6379") // server setup
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[Server] Listening on Port: 6379")

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for { // server loop
		resp := NewResp(conn) // read from conn

		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(value)

		writer := NewWriter(conn)
		writer.Write(Value{typ: "string", str: "OK"})
	}
}
