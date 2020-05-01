// Based on https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf
package chord

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/pkg/helpers"
	"google.golang.org/grpc"
)

// Node chord node
type Node struct {
	Identifier       [sha256.Size]byte // hashed value of node
	IP               string
	Port             int
	Successor        *Node // next node
	Predecessor      *Node // previous node
	FingerTable      map[int]*Node
	FingerFixerIndex int // to use in fixFinger
	M                int // number of bits in identifier
	connPool         map[string]*grpc.ClientConn
	mutex            sync.RWMutex
}

// NewNode create new node
func NewNode(ip string, port int) *Node {
	id := sha256.Sum256([]byte(ip + ":" + strconv.FormatInt(int64(port), 10)))
	newNode := &Node{
		Identifier:       id,
		IP:               ip,
		Port:             port,
		FingerTable:      make(map[int]*Node),
		FingerFixerIndex: 0,
		M:                sha256.Size * 8,
		connPool:         make(map[string]*grpc.ClientConn),
	}
	newNode.Successor = newNode
	newNode.Predecessor = nil
	return newNode
}

// Join throw first node
func (n *Node) Join(remote *Node) {
	client := n.Connect(remote)
	successor, err := client.FindSuccessor(context.Background(), &pb.Lookup{Key: n.Identifier[:]})
	if err != nil {
		fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
		return
	}
	fmt.Printf("Join: got successor %s:%d! \n", successor.IP, successor.Port)
	n.Predecessor = nil
	n.Successor = NewNode(successor.IP, int(successor.Port))

	n.mutex.Lock()
	n.FingerTable[1] = n.Successor
	n.mutex.Unlock()
	n.RemoteSuccessorNotify() // update new successor's predecessor
}

// Notify update predecessor
// is being called periodically by predecessor or new node
func (n *Node) Notify(node *Node) bool {
	// (n.predecessor is nil or node ∈ (n.predecessor, n))
	if n.Predecessor == nil {
		n.Predecessor = node
		if n.Successor.Identifier == n.Identifier { // bootstrap node
			n.Successor = node
			n.RemoteSuccessorNotify() // notify successor
		}
		return true
	}
	if helpers.Between(node.Identifier, n.Predecessor.Identifier, n.Identifier) {
		n.Predecessor = node
		return true
	}
	return false
}

// FindSuccessor find the closest node to the given identifier
func (n *Node) FindSuccessor(identifier [sha256.Size]byte) *Node {
	// fmt.Printf("FindSuccessor: start looking for key %x \n", identifier)
	if n.Successor.Identifier == n.Identifier {
		return n
	}
	// id ∈ (n, successor]
	if helpers.BetweenR(identifier, n.Identifier, n.Successor.Identifier) {
		return n.Successor
	}
	nextNode := n.closestPrecedingNode(identifier)
	if nextNode.Identifier == n.Identifier { // current node is the only node in figer table
		return nextNode
	}
	return n.FindRemoteSuccessor(nextNode, identifier)
}

func (n *Node) closestPrecedingNode(identifier [sha256.Size]byte) *Node {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	for m := len(n.FingerTable); m > 0; m-- {
		// finger[i] ∈ (n, id)
		//fmt.Printf("closestPrecedingNode: checking figertable[%d] => id, %x", m, n.FingerTable[m].Identifier)
		if helpers.Between(n.FingerTable[m].Identifier, n.Identifier, identifier) {
			return n.FingerTable[m]
		}
	}
	return n
}

// FindRemoteSuccessor find closest node to the given key in remote node
func (n *Node) FindRemoteSuccessor(remote *Node, key [sha256.Size]byte) *Node {
	client := n.Connect(remote)
	successor, err := client.FindSuccessor(context.Background(), &pb.Lookup{Key: key[:]})
	chordNode := NewNode(successor.IP, int(successor.Port))
	if err != nil {
		fmt.Printf("There is no remote successor from: %s:%d \n", remote.IP, remote.Port)
	} else {
		//fmt.Printf("Remote node has found successor %s:%d! \n", successor.IP, successor.Port)
	}
	return chordNode
}

// GetRemotePredecessor get remote nodes predecessor
func (n *Node) GetRemotePredecessor(remote *Node) *Node {
	// prepare kind of timeout to replace disconnected nodes
	client := n.Connect(remote)
	predecessor, err := client.GetPredecessor(context.Background(), new(empty.Empty), grpc.WaitForReady(true))
	if err != nil {
		fmt.Printf("There is no remote predecessor ERROR: %+v \n", err)
	}
	chordNode := NewNode(predecessor.IP, int(predecessor.Port))
	if err != nil {
		fmt.Printf("There is no remote predecessor from: %s:%d \n", remote.IP, remote.Port)
	} else {
		//fmt.Printf("Remote node has found predecessor %s:%d! \n", predecessor.IP, predecessor.Port)
	}
	return chordNode
}

// RemoteSuccessorNotify notify remote node (update it's predecessur)
func (n *Node) RemoteSuccessorNotify() (bool, error) {
	client := n.Connect(n.Successor) // connect to the successor
	node := &pb.Node{
		IP:   n.IP,
		Port: int32(n.Port),
		Key:  n.Identifier[:], // FIXME remove identifier from grpc
	}
	result, err := client.Notify(context.Background(), node)
	if err != nil {
		fmt.Printf("Error notifying successo: %s:%d \n", n.Successor.IP, n.Successor.Port)
		return false, err
	}
	// fmt.Printf("Remote node has notified %+v! \n", result)
	return result.Value, err
}

// Stabilize keep successor and predecessor updated
// Runs periodically
func (n *Node) Stabilize() {
	// If node has no successor yet, ignore
	if n.Successor == nil {
		return
	}
	remotePredecessor := n.GetRemotePredecessor(n.Successor)
	if remotePredecessor.Identifier == n.Identifier {
		return
	}
	// means successor's predececcor is changed

	// if pred(succ) ∈ (n, succ)
	if helpers.BetweenR(remotePredecessor.Identifier, n.Identifier, n.Successor.Identifier) {
		n.Successor = remotePredecessor
		n.RemoteSuccessorNotify()
	}
}

// CheckPredecessor keeps predecessor uptodate
// Runs periodically
func (n *Node) CheckPredecessor() {
	if n.Predecessor != nil {
		if !n.Ping(n.Predecessor) { // predecessor disconnected !
			n.Predecessor = nil // set nil to be able to update predecessor by notify
		}
	}
}

// FixFingers refreshes finger table entities
// Runs periodically
func (n *Node) FixFingers() {
	if n.Successor == nil {
		return
	}
	n.FingerFixerIndex++
	if n.FingerFixerIndex > 5 {
		n.FingerFixerIndex = 1
	}
	// finger[k] = (n + 2 ** k-1) Mod M

	meint := new(big.Int)
	meint.SetBytes(n.Identifier[:])

	baseint := new(big.Int)
	baseint.SetUint64(2)

	powint := new(big.Int)
	powint.SetInt64(int64(n.FingerFixerIndex - 1))

	var biggest [sha256.Size + 1]byte
	for i := range biggest {
		biggest[i] = 255
	}

	tmp := new(big.Int)
	tmp.SetInt64(1)

	modint := new(big.Int)
	modint.SetBytes(biggest[:sha256.Size])
	modint.Add(modint, tmp)

	target := new(big.Int)
	target.Exp(baseint, powint, modint)
	target.Add(meint, target)
	target.Mod(target, modint)

	bytes := target.Bytes()
	diff := sha256.Size - len(bytes)
	if diff > 0 {
		tmp := make([]byte, sha256.Size)
		//pad with zeros
		for i := 0; i < diff; i++ {
			tmp[i] = 0
		}
		for i := diff; i < sha256.Size; i++ {
			tmp[i] = bytes[i-diff]
		}
		bytes = tmp
	}
	var identifier [sha256.Size]byte
	copy(identifier[:sha256.Size], bytes[:sha256.Size])
	findSuccessor := n.FindSuccessor(identifier)
	n.mutex.Lock()
	n.FingerTable[n.FingerFixerIndex] = findSuccessor
	n.mutex.Unlock()
	// if n.FingerFixerIndex == 1 { // first item in fingertable is immediate successor
	// 	n.Successor = n.FingerTable[n.FingerFixerIndex]
	// }
}

// Debug prints successor,predecessor
// Runs periodically
func (n *Node) Debug() {
	fmt.Printf("Current Node: %s:%d:%x\n", n.IP, n.Port, n.Identifier)
	if n.Successor != nil {
		fmt.Printf("Current Node Successor: %s:%d:%x\n", n.Successor.IP, n.Successor.Port, n.Successor.Identifier)
	}
	if n.Predecessor != nil {
		fmt.Printf("Current Node Predecessor: %s:%d:%x\n\n", n.Predecessor.IP, n.Predecessor.Port, n.Predecessor.Identifier)
	}
}

// Connect grpc connect to remote node
// FIXME should be cached
func (n *Node) Connect(remote *Node) pb.ChordClient {
	addr := remote.IP + ":" + strconv.FormatInt(int64(remote.Port), 10)
	if n.connPool[addr] == nil {
		conn, _ := grpc.Dial(addr, grpc.WithInsecure())
		n.connPool[addr] = conn
	}
	client := pb.NewChordClient(n.connPool[addr])
	return client
}

// Ping check if remote port is open
// FIXME should be cached
func (n *Node) Ping(remote *Node) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(remote.IP, strconv.FormatInt(int64(remote.Port), 10)), timeout)
	if err != nil {
		//fmt.Printf("Ping to %s:%d error:%v", remote.IP, remote.Port, err)
		return false
	}
	if conn != nil {
		defer conn.Close()
		//fmt.Printf("Ping successful %s:%d", remote.IP, remote.Port)
		return true
	}
	return false
}
