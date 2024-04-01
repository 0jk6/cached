package main

import (
	"fmt"
	"os"

	cacheserver "cached.0jk6.github.io/internal"
	raft "cached.0jk6.github.io/internal/raft"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("-----------------------------------")
		fmt.Println("Run the key-value server: ./main <port>")
		fmt.Println("Run the raft cluster: ./main <node1> <node2> <node3> ...")
		fmt.Println("-----------------------------------")
		return
	} else if len(os.Args) > 2 {
		raft := raft.Raft{
			LeaderId: -1,
		}
		fmt.Printf("Raft cluster started with %d nodes\n", len(os.Args)-1)
		raft.CreateCluster(os.Args[1:])
		raft.ScheduleElection(10)
	} else {
		port := ":" + os.Args[1]

		//tcp server
		// tcpServer := cacheserver.TCPServer{
		// 	LogFile: "./wal.log",
		// }
		// tcpServer.Listen(port)

		// http server
		httpServer := cacheserver.HTTPServer{
			LogFile: "./wal.log",
		}

		httpServer.Listen(port)
	}

}
