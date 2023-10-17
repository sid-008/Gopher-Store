package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[Server] Listening on Port: 6379")

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for {
		buf := make([]byte, 1024)

		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Error reading from client: ", err.Error())
			os.Exit(1)
		}
		_, err = conn.Write([]byte("+OK\r\n"))
		if err != nil {
			log.Println(err)
		}

	}
}
