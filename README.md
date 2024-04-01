# cached

A distributed key-value store using Raft algorithm.

Replicates key-value server's data to all the other nodes in the cluster. Elects a leader based on the commit indices of the nodes

### How to run

`make build` or `go build cmd/main.go`

Run the key-value server: `./main <port>`

Run the raft cluster: `./main <node1> <node2> <node3> ...`

example:

run `./main 3000`, `./main 3001`, `./main 3002` in 3 different terminals

run the raft cluster using: `./main http://localhost:3000 http://localhost:3001 http://localhost:3002`


###### Note
This implementation uses basic ideas of Raft, just to make myself comfortable with distributed systems. It's probably the worst implementation of the Raft algorithm, but it works!!!