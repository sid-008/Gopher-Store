package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":6379") // server setup
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[Server] Listening on Port: 6379")

	aof, err := NewAOF("./database.aof")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer aof.Close()

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		if value.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(value.array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk) // redis commands
		args := value.array[1:]

		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		if command == "SET" {
			aof.Write(value)
		}

		result := handler(args)
		writer.Write(result)
	}
}
