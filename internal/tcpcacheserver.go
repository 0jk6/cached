package cacheserver

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type TCPServer struct {
	LogFile string
}

func (server *TCPServer) ListenTCP(port string) {
	cache := &Cache{
		LogFile: server.LogFile,
	}

	listener, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("listening on port", port)

	//accept incoming requests
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("client", conn.RemoteAddr().String(), "connected")
			go handleClient(conn, cache)
		}
	}
}

func handleClient(conn net.Conn, cache *Cache) {
	defer conn.Close()

	//read incoming data into a buffer
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			//if we use log.Fatal(err), the server will terminate because of EOF
			if err.Error() != "EOF" {
				fmt.Println("Error reading from connection:", err)
			}
			return
		}

		data := string(buffer[:n])
		resp, closeConn := processData(data, cache)

		if closeConn {
			conn.Write([]byte(resp))
			fmt.Println("client", conn.RemoteAddr().String(), "disconnected")
			return
		} else {
			conn.Write([]byte(resp))
		}
	}
}

func processData(data string, cache *Cache) (string, bool) {
	//split data string
	tokens := strings.Fields(data)
	resp := ""
	closeConn := false

	if len(tokens) == 0 {
		return resp, closeConn
	}

	tokens[0] = strings.ToLower(tokens[0])

	if tokens[0] == "get" {
		if len(tokens) >= 2 {
			return cache.getData(tokens[1]), false
		} else {
			return "not enough args\r\n", false
		}
	} else if tokens[0] == "set" {
		if len(tokens) >= 3 {
			return cache.storeData(tokens[1], tokens[2]), false
		} else {
			return "not enough args\r\n", false
		}
	} else if tokens[0] == "del" {
		if len(tokens) >= 2 {
			return cache.deleteData(tokens[1]), false
		} else {
			return "not enough args\r\n", false
		}
	} else if tokens[0] == "exit" {
		resp = "exiting...\r\n"
		closeConn = true
	} else {
		if tokens[0] == "" {
			resp = ""
		} else {
			resp = fmt.Sprintf("unknown: %v\r\n", tokens[0])
		}
	}

	return resp, closeConn
}
