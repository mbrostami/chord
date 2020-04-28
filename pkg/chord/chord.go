package chord

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/pkg/helpers"
	"google.golang.org/grpc"
)

// Node chord node
type Node struct {
	Identifier  [sha256.Size]byte // hashed value of node
	IP          string
	Port        int
	Successor   *Node // next node
	Predecessor *Node // previous node
	FingerTable map[int]*Node
}

// NewNode create new node
func NewNode(ip string, port int) *Node {
	id := sha256.Sum256([]byte(ip + ":" + strconv.FormatInt(int64(port), 10)))
	return &Node{
		Identifier:  id,
		IP:          ip,
		Port:        port,
		FingerTable: make(map[int]*Node),
	}
}

// FindSuccessor find the closest node to the given identifier
func (n *Node) FindSuccessor(identifier [sha256.Size]byte) *Node {
	if len(n.FingerTable) == 0 {
		fmt.Printf("FindSuccessor: fingerTable has no successor. Looking for %x, returned self: %x \n", identifier, n.Identifier)
		return n // if fingertable is empty, return self
	}
	successorNode := n.FingerTable[1]
	fmt.Printf("FindSuccessor: start looking for key %x, localid: %+v \n", identifier, n.FingerTable)
	// id ∈ (n, successor]
	if helpers.OpLTE(identifier, successorNode.Identifier) {
		if helpers.OpGT(identifier, n.Identifier) {
			fmt.Printf("FindSuccessor: found local successor hash: %x \n", successorNode.Identifier)
			return successorNode
		}
	}
	nextNode := n.closestPrecedingNode(identifier)
	return n.FindRemoteSuccessor(nextNode, identifier)
}

func (n *Node) closestPrecedingNode(identifier [sha256.Size]byte) *Node {
	for m := len(n.FingerTable); m > 0; m-- {
		// finger[i] ∈ (n, id)
		if helpers.OpGT(n.FingerTable[m].Identifier, n.Identifier) && helpers.OpLT(n.FingerTable[m].Identifier, identifier) {
			fmt.Printf("Lookup: local finger[%d] hash: %x \n", m, n.FingerTable[m].Identifier)
			return n.FingerTable[m]
		}
	}
	fmt.Print("Lookup: local finger empty \n")
	// FIXME prevent infinite call
	return n
}

// FindRemoteSuccessor find closest node to the given key in remote node
func (n *Node) FindRemoteSuccessor(remote *Node, key [sha256.Size]byte) *Node {
	client := n.Connect(remote)
	successor, err := client.FindSuccessor(context.Background(), &pb.Lookup{Key: key[:]})
	chordNode := &Node{}
	if err != nil {
		fmt.Printf("There is no remote successor from: %s:%d \n", remote.IP, remote.Port)
		return chordNode
	}
	var id [sha256.Size]byte
	copy(id[:], successor.Key[:sha256.Size])
	chordNode.Identifier = id
	chordNode.IP = successor.IP
	chordNode.Port = int(successor.Port)
	fmt.Printf("Remote node has found successor %+v! \n", successor)
	return chordNode
}

// Join throw first node
func (n *Node) Join(remote *Node) {
	client := n.Connect(remote)
	predecessor, err := client.GetPredecessor(context.Background(), new(empty.Empty))
	if err != nil {
		fmt.Printf("There is no predecessor from: %s:%d \n", remote.IP, remote.Port)
		return
	}
	if predecessor.Port < 1 {
		fmt.Printf("Remote node has no predecessor! \n")
		// remote is the only node
		n.Predecessor = remote
	} else {
		var id [sha256.Size]byte
		copy(id[:], predecessor.Key[:sha256.Size])
		n.Predecessor = &Node{
			Identifier: id,
			IP:         predecessor.IP,
			Port:       int(predecessor.Port),
		}
		fmt.Printf("Remote node has predecessor %+v! \n", predecessor)
	}
	n.Successor = remote
	n.FingerTable[1] = remote
}

func (n *Node) Connect(remote *Node) pb.ChordClient {
	conn, _ := grpc.Dial(remote.IP+":"+strconv.FormatInt(int64(remote.Port), 10), grpc.WithInsecure())
	client := pb.NewChordClient(conn)
	return client
}

// // NewChord r = desired redoundancy
// func NewChord(ipaddr string, port int) *Node {
// 	// FIXME id should be hash value of address
// 	id := sha256.Sum256([]byte(ipaddr + ":" + string(port)))
// 	node := &Node{
// 		Identifier:  id,
// 		Ip:          ipaddr,
// 		Port:        port,
// 		FingerTable: make(map[int]*Node),
// 	}
// 	return node
// }

// func (n *Node) RegisterBootstrapNode(ip string, port int) {
// 	n.BootstrapNode.ip = ip
// 	n.BootstrapNode.Port = port
// }

// func (n *Node) Listen(res chan string, wait *sync.WaitGroup) {
// 	laddr := new(net.TCPAddr)
// 	laddr.IP = net.ParseIP(n.ip)
// 	laddr.Port = n.Port
// 	// go func() {
// 	listener, _ := net.ListenTCP("tcp", laddr) // listen in background
// 	wait.Done()
// 	defer fmt.Printf("No longer listening...\n")
// 	for {
// 		if conn, err := listener.AcceptTCP(); err == nil {
// 			err = conn.SetDeadline(time.Now().Add(3 * time.Minute))
// 			var data []byte
// 			defer conn.Close()
// 			nread, _ := conn.Read(data)
// 			fmt.Printf("Read %v \n", nread)
// 			msg := string(data[:nread])
// 			functionName := strings.Split(msg, ":")[0]
// 			arg := strings.Split(msg, ":")[1]
// 			var response string
// 			if functionName == "findSuccessor" {
// 				response = n.FindSuccessor(arg)
// 			}
// 			nwrite, _ := conn.Write([]byte(response))
// 			fmt.Printf("Write %v: Response: %s", nwrite, response)
// 			res <- (n.ip + ":" + string(n.Port))
// 		} else {
// 			continue
// 		}
// 	}
// }

// // AddRemoteNode idx = 1 is successor
// func (n *Node) Join() {
// 	// finger[i] = successor(n+(2^i-1) mod 2^m)
// 	return n.sendRemoteMsg(n.BootstrapNode.ip, n.BootstrapNode.Port, "findSuccessor:"+key)
// }

// func (n *Node) AddRemoteNode(idx int, ripaddr string, rport int) *Node {
// 	id := sha256.Sum256([]byte(ripaddr + ":" + string(rport)))
// 	rNode := &Node{
// 		identifier:  id,
// 		ip:          ripaddr,
// 		Port:        rport,
// 		FingerTable: make(map[int]*Node),
// 	}
// 	// finger[k] = (n + 2^k-1) mod 2^m , 1 <= k <= m
// 	n.FingerTable[idx] = rNode
// 	return rNode
// }

// func (n *Node) FindSuccessor(key string) string {
// 	hashedKey := sha256.Sum256([]byte(key))
// 	successorNode := n.FingerTable[1]
// 	fmt.Printf("Lookup: start looking for key %s, hash: %x \n", key, hashedKey)
// 	// id ∈ (n, successor]
// 	if lte(hashedKey, successorNode.identifier) && gt(hashedKey, n.identifier) {
// 		fmt.Printf("Lookup: local successor hash: %x \n", successorNode.identifier)
// 		return successorNode.ip + ":" + strconv.FormatInt(int64(successorNode.Port), 10)
// 	}
// 	nextNode := n.closestPrecedingNode(hashedKey)
// 	return n.remoteFindSuccessor(nextNode, key)
// }

// func (n *Node) closestPrecedingNode(identifier [mSize]byte) *Node {
// 	for m := len(n.FingerTable); m > 0; m-- {
// 		// finger[i] ∈ (n, id)
// 		if gt(n.FingerTable[m].identifier, n.identifier) && lt(n.FingerTable[m].identifier, identifier) {
// 			fmt.Printf("Lookup: local finger[%d] hash: %x \n", m, n.FingerTable[m].identifier)
// 			return n.FingerTable[m]
// 		}
// 	}
// 	fmt.Print("Lookup: local finger empty \n")
// 	// FIXME prevent infinite call
// 	return n
// }

// func (n *Node) remoteFindSuccessor(remote *Node, key string) string {
// 	return n.sendRemoteMsg(remote.ip, remote.Port, "findSuccessor:"+key)
// }

// func (n *Node) sendRemoteMsg(ip string, port int, message string) string {
// 	laddr := new(net.TCPAddr)
// 	laddr.IP = net.ParseIP(n.ip) // should be current running node
// 	laddr.Port = 0
// 	raddr := new(net.TCPAddr)
// 	raddr.IP = net.ParseIP(ip)
// 	raddr.Port = port
// 	newconn, _ := net.DialTCP("tcp", laddr, raddr)
// 	conn := *newconn
// 	defer conn.Close()
// 	msg := []byte(message)
// 	_, err := conn.Write(msg)
// 	if err != nil {
// 		return "ERROR"
// 	}
// 	var reply []byte
// 	res, err := conn.Read(reply)
// 	if err != nil {
// 		return "ERROR"
// 	}
// 	reply = reply[:res]
// 	return string(reply)
// }

// func (c *Chord) Insert(key []byte, value []byte) (bool, error) {
// 	var result bool
// 	// insert key value, in r carefully chosen nodes (sync)
// 	return result, nil
// }

// func (c *Chord) Update(key []byte, value []byte) (bool, error) {
// 	var result bool
// 	// Chord allows updates to a key/value binding,
// 	// but currently only by the originator of the key.
// 	// This restriction simplifies the mechanisms required to provide
// 	// correct update semantics when network partitions heal
// 	//
// 	return result, nil
// }

// func (c *Chord) Lookup(key []byte) ([]byte, error) {
// 	var value []byte
// 	return value, nil
// }

// func (c *Chord) Join(node string) (bool, error) {
// 	var result bool
// 	return result, nil
// }

// func (c *Chord) Leave() (bool, error) {
// 	var result bool
// 	return result, nil
// }

// Stabilization If the underlying network connecting Chord servers suffers a partition,
// the servers in each partition communicate with each other
// to reorganize the overlay within the partition, assuring that there will be eventually r
// distinct nodes storing each binding.
// When partitions heal, a stabilization protocol assures that there will be exactly r
// distributed locations for any binding in any connected partition
// func (c *Chord) Stabilization() {
// 	// c.r
// }
