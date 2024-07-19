package raft

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Node is the key-value server
type RaftNode struct {
	id          int
	serverAddr  string
	votedFor    int
	term        int
	state       string   //follower, candidate, leader
	commitIndex int      //based on this commit id, leader election will start
	LogEntry    []string //sync this by pulling them from the leader node
}

// Raft is the main object with which we will be creating a cluster
type Raft struct {
	mu       sync.Mutex
	term     int
	LeaderId int
	nodes    []*RaftNode
}

// this will be used while creating the raft cluster
func createNode(nodeId int, serverAddr string) *RaftNode {
	return &RaftNode{
		id:          nodeId,
		serverAddr:  serverAddr,
		votedFor:    -1,
		term:        0,
		state:       "follower",
		commitIndex: 0,
	}
}

// a server can invoke this function to request a vote
func (node *RaftNode) requestForVote(serverId, commitIndex int) bool {
	//vote should be given based on the serverId's commit index
	if commitIndex > node.commitIndex {
		fmt.Println("Node", node.serverAddr, "voted for node", serverId)
		node.votedFor = serverId
	} else {
		//vote yourself
		node.votedFor = node.id
	}

	return node.votedFor == serverId
}

// create a cluster
func (r *Raft) CreateCluster(serverAddrs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nodes = make([]*RaftNode, len(serverAddrs))

	for i, serverAddr := range serverAddrs {
		r.nodes[i] = createNode(i, serverAddr)
	}
}

func (r *Raft) StartElection(nodeId int) {
	fmt.Printf("Node %d has started the election\n", nodeId)

	r.term++
	r.nodes[nodeId].state = "candidate"
	r.nodes[nodeId].votedFor = nodeId

	numVotes := 1

	var wg sync.WaitGroup

	//start asking for votes from each node
	for _, node := range r.nodes {
		if node.id != nodeId {
			wg.Add(1)
			go func(nodeId int, node *RaftNode, commitIndex int) {
				defer wg.Done()
				ok := node.requestForVote(nodeId, commitIndex)

				if ok {
					r.mu.Lock()
					numVotes++
					r.mu.Unlock()
				}

			}(nodeId, node, r.nodes[nodeId].commitIndex)
		}
	}
	wg.Wait()

	//check the votes and elect the leader
	if numVotes > len(r.nodes)/2 {
		r.LeaderId = nodeId
		r.nodes[nodeId].state = "leader"

		fmt.Printf("Node %d has won the election\n", nodeId)

		for _, node := range r.nodes {
			if node.id != nodeId {
				node.state = "follower"
			}
		}

		//TODO: start log replication
		fmt.Println("replicating logs to all the key-value servers")
		r.replicateLogs()
	} else {
		fmt.Printf("node %d has lost the election\n", nodeId)
	}
}

// for now let's run this function every 10 seconds and check the commit indices of all the nodes
// and pull the commit index of highest node, based on that elect the leader
func (r *Raft) ScheduleElection(scheduleInterval int) {

	fmt.Println("Scheduling election every", scheduleInterval, "seconds...")
	for {

		// make a http call to all the nodes and get the commit index
		r.getCommitIds()

		maxCommitIndex := 0
		maxCommitIndexNodeId := -1

		for _, node := range r.nodes {
			//get the commit index of other nodes using http (this should be an rpc)
			if node.commitIndex > maxCommitIndex {
				maxCommitIndex = node.commitIndex
				maxCommitIndexNodeId = node.id
			}
		}

		if maxCommitIndexNodeId != -1 && r.LeaderId != maxCommitIndexNodeId {
			fmt.Printf("Node %d has the highest commit index... starting election\n", maxCommitIndexNodeId)
			r.StartElection(maxCommitIndexNodeId)
		}

		time.Sleep(time.Duration(scheduleInterval) * time.Second)
	}

}

func (r *Raft) getCommitIds() {
	for _, node := range r.nodes {
		serverAddr := node.serverAddr
		//make a http call to get the commit index
		resp, err := http.Get(serverAddr + "/commit")

		if err != nil {
			fmt.Println("err while getting commit index", err)
		} else {
			data, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			if err != nil {
				fmt.Println("err while decoding data")
			} else {
				dataStr := string(data)
				r.mu.Lock()
				// r.nodes[node.id].commitIndex, _ = strconv.Atoi(dataStr)
				node.commitIndex, _ = strconv.Atoi(dataStr)
				r.mu.Unlock()
			}
		}
	}
}

func (r *Raft) replicateLogs() {
	//replicate logs to all the key-value servers
	//leader will send it's current logs to all the followers
	//followers will append the logs to their logs
	//once all the followers have appended the logs, leader will commit the logs
	//once the logs are committed, leader will send the commit index to all the followers
	//followers will update their commit index
	//once the commit index is updated, leader will send the response to the client
	//client will get the response and send the next request
	//this is how the raft algorithm works

	// check if the current node is the leader
	if r.nodes[r.LeaderId].state != "leader" {
		fmt.Println("Only the leader can replicate logs")
		return
	}

	//---------------------LOG REPLICATION---------------------
	//first get the commitIndex and LogEntry of the leader node
	//then get the commitIndex and LogEntry of follower nodes
	//sync the difference between them
	//follower node's final log entry will be as follows

	leaderAddr := r.nodes[r.LeaderId].serverAddr

	resp, err := http.Get(leaderAddr + "/logentry")
	if err != nil {
		fmt.Println("error fetching data from leader", err)
		return
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("error while reading data", err)
	}

	//following is the base64 encoded log entry of the leader node
	leaderLogEntryBase64 := string(data)
	// fmt.Println(leaderLogEntryBase64)

	resp, err = http.Get(leaderAddr + "/commit")
	if err != nil {
		fmt.Println("error fetching data from leader", err)
		return
	}

	data, err = io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("error while reading data", err)
	}

	leaderCommitIndex := string(data)
	fmt.Println(leaderCommitIndex)

	leaderLogEntryBytes, _ := base64.StdEncoding.DecodeString(leaderLogEntryBase64)
	leaderLogEntry := string(leaderLogEntryBytes)
	// fmt.Println(leaderLogEntry)

	leaderLogEntrySplits := strings.Split(leaderLogEntry, "\r\n")

	for _, node := range r.nodes {
		// fmt.Println(node)
		//get commit indices
		commitIndex := node.commitIndex
		if commitIndex < r.nodes[r.LeaderId].commitIndex {
			// follower node's new log entry will be as follows
			//leaderLogEntrySplits[node.commitIndex:]
			node.appendEntries(leaderLogEntrySplits[node.commitIndex:])
		}
	}
}

func (node *RaftNode) appendEntries(entriesToBeAppended []string) {
	fmt.Println("appending entries in the node", node.id)
	// fmt.Println(entriesToBeAppended)
	entriesToBeAppendedBase64 := base64.StdEncoding.EncodeToString([]byte(strings.Join(entriesToBeAppended, ",")))

	nodeAddr := node.serverAddr + "/appendentries/" + entriesToBeAppendedBase64
	resp, err := http.Get(nodeAddr)

	if err != nil {
		fmt.Println("error while appending entries in node", node.serverAddr, err)
		return
	}

	data, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Println(string(data))
}
