package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handler(conn net.Conn) {
	b := make([]byte, 512)

	_, err := conn.Read(b[0:])

	if err == io.EOF {
		log.Println("Client disconnected")
	}

	if err != nil {
		log.Println("Unexpected error")
	}

	response := []byte("HTTP/1.1 200 OK\nContent-Type: image/jpeg\nContent-Length: 3\nAccess-Control-Allow-Origin: *\n\nimg")

	fmt.Println("Read:")
	fmt.Println(string(b))

	fmt.Println("Writing...")
	if _, err := conn.Write(response); err != nil {
		log.Fatalln("Unable to write data")
	}

	fmt.Println("Closing...")
	conn.Close()
}

func main() {
	listener, err := net.Listen("tcp", ":4001")

	if err != nil {
		log.Println("Unable to bind to port")
	}

	log.Println("Listening to 0.0.0.0:4001")

	for {
		conn, err := listener.Accept()
		log.Println("Received connection")

		if err != nil {
			log.Println("Unable to accept connection")
		}

		go handler(conn)
	}
}
