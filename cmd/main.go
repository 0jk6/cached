package main

import (
	"fmt"
	"os"

	cacheserver "cached.0jk6.github.io/internal"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("-------------------------------------")
		fmt.Println("Usage:", os.Args[0], "[port-number]")
		fmt.Println("Example:", os.Args[0], "3000")
		fmt.Println("-------------------------------------")
		return
	}

	port := ":" + os.Args[1]

	server := cacheserver.TCPServer{
		LogFile: "./wal.log",
	}

	server.ListenTCP(port)
}
