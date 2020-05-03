// Based on https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf

package chord

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/pkg/helpers"
	"google.golang.org/grpc"
)

// MSIZE is the number of bits in identifier
// in fact only O(log n) are distinct
// ref D
const MSIZE int = sha256.Size * 8

// RSIZE is the number of records in successor list (max = (log n))
// ref E.3
const RSIZE int = 5

// Chord chord ring
// ref B
type Chord struct {
	Successor        *Node // next node
	Predecessor      *Node // previous node
	SuccessorList    map[int]*Node
	FingerTable      map[int]*Node // ref D
	FingerFixerIndex int           // to use in fixFinger
	connectionPool   map[string]*grpc.ClientConn
	mutex            sync.RWMutex
	m                int // ref D - MSIZE
	r                int // ref E.3 - RSIZE
	Node             Node
	fingetTableDebug map[int]string // FIXME remove this one, it's for debug
}

// NewChord create new node
func NewChord(ip string, port int) *Chord {
	chord := &Chord{
		SuccessorList:    make(map[int]*Node),
		FingerTable:      make(map[int]*Node),
		FingerFixerIndex: 0,
		m:                MSIZE,
		r:                RSIZE,
		connectionPool:   make(map[string]*grpc.ClientConn),
		fingetTableDebug: make(map[int]string),
	}

	chord.Node.IP = ip
	chord.Node.Port = port
	chord.Node.Identifier = helpers.Hash(ip, port)

	chord.Successor = &chord.Node
	chord.Predecessor = nil
	return chord
}

// Join throw first node
func (c *Chord) Join(remote *Node) {
	client := c.Connect(remote)
	successor, err := client.FindSuccessor(context.Background(), &pb.Lookup{Key: c.Node.Identifier[:]})
	if err != nil {
		fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
		return
	}
	fmt.Printf("Join: got successor %s:%d! \n", successor.IP, successor.Port)
	c.Predecessor = nil
	c.Successor = &NewChord(successor.IP, int(successor.Port)).Node

	c.mutex.Lock()
	c.FingerTable[1] = c.Successor
	c.mutex.Unlock()
	c.RemoteSuccessorNotify() // update new successor's predecessor
}

// Notify update predecessor
// is being called periodically by predecessor or new node
func (c *Chord) Notify(node *Node) bool {
	// (c.predecessor is nil or node ∈ (c.predecessor, n))
	if c.Predecessor == nil {
		c.Predecessor = node
		if c.Successor.Identifier == c.Node.Identifier { // bootstrap node
			c.Successor = node
			c.RemoteSuccessorNotify() // notify successor
		}
		return true
	}
	if helpers.Between(node.Identifier, c.Predecessor.Identifier, c.Node.Identifier) {
		c.Predecessor = node
		return true
	}
	return false
}

// FindSuccessor find the closest node to the given identifier
// ref D
func (c *Chord) FindSuccessor(identifier [sha256.Size]byte) *Node {
	// fmt.Printf("FindSuccessor: start looking for key %x \n", identifier)
	if c.Successor.Identifier == c.Node.Identifier {
		return &c.Node
	}
	// id ∈ (n, successor]
	if helpers.BetweenR(identifier, c.Node.Identifier, c.Successor.Identifier) {
		return c.Successor
	}
	nextNode := c.closestPrecedingNode(identifier)
	if nextNode.Identifier == c.Node.Identifier { // current node is the only node in figer table
		return nextNode
	}
	return c.FindRemoteSuccessor(nextNode, identifier)
}

// ref D
func (c *Chord) closestPrecedingNode(identifier [sha256.Size]byte) *Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for m := len(c.FingerTable); m > 0; m-- {
		// finger[i] ∈ (n, id)
		// fmt.Printf("closestPrecedingNode: checking figertable[%d] => id, %x", m, c.FingerTable[m].Identifier)
		if helpers.Between(c.FingerTable[m].Identifier, c.Node.Identifier, identifier) {
			return c.FingerTable[m]
		}
	}
	return &c.Node
}

// FindRemoteSuccessor find closest node to the given key in remote node
func (c *Chord) FindRemoteSuccessor(remote *Node, key [sha256.Size]byte) *Node {
	client := c.Connect(remote)
	successor, err := client.FindSuccessor(context.Background(), &pb.Lookup{Key: key[:]})
	chordNode := &NewChord(successor.IP, int(successor.Port)).Node
	if err != nil {
		fmt.Printf("There is no remote successor from: %s:%d \n", remote.IP, remote.Port)
	} else {
		//fmt.Printf("Remote node has found successor %s:%d! \n", successor.IP, successor.Port)
	}
	return chordNode
}

// GetPredecessor return predecessor of current node, or replace current predecessor with caller if applicable
// ref E.1
func (c *Chord) GetPredecessor(caller *Node) *Node {
	if c.Predecessor != nil {
		// extension on chord
		if helpers.Between(caller.Identifier, c.Predecessor.Identifier, c.Node.Identifier) {
			c.Predecessor = caller // update predecessor if caller is closer than predecessor
		}
		return c.Predecessor
	}
	return &c.Node
}

// GetSuccessorList returns unsorted successor list
// ref E.3
func (c *Chord) GetSuccessorList() map[int]*Node {
	return c.SuccessorList
}

// GetRemoteStablierData successor's (successor list and predecessor)
// to prevent duplicate rpc call, we get both together
// ref E.3
func (c *Chord) GetRemoteStablierData(remote *Node) (*Node, map[int]*Node) {
	// prepare kind of timeout to replace disconnected nodes
	client := c.Connect(remote)

	stablizerData, err := client.GetStablizerData(context.Background(), c.Node.GrpcNode(), grpc.WaitForReady(true))
	if err != nil {
		fmt.Printf("There is no remote predecessor ERROR: %+v \n", err)
	}
	predecessor := stablizerData.Predecessor
	chordNode := &NewChord(predecessor.IP, int(predecessor.Port)).Node

	nodes := c.prepareSuccessorList(stablizerData.Nodes)
	return chordNode, nodes
}

// prepareSuccessorList get remote successor list
func (c *Chord) prepareSuccessorList(pbNodes *pb.Nodes) map[int]*Node {
	successorList := pbNodes
	nodes := make(map[int]*Node)
	nodes[0] = c.Successor // replace first item with successor itself
	// fmt.Printf("successorList starting to update \n")
	index := 1
	for i := 0; i < len(successorList.Nodes); i++ {
		if len(nodes) >= c.r { // prevent overloading successorlist (max(r)=(log N)) ref E.3
			break
		}
		chorNode := &NewChord(successorList.Nodes[i].IP, int(successorList.Nodes[i].Port)).Node
		// ignore same nodes
		if chorNode.Identifier == c.Node.Identifier {
			continue
		}
		nodes[index] = chorNode
		// fmt.Printf("successorList added: %d ::: %x\n", index, chorNode.Identifier)
		index++
	}
	return nodes
}

// RemoteSuccessorNotify notify remote node (update it's predecessur)
func (c *Chord) RemoteSuccessorNotify() (bool, error) {
	client := c.Connect(c.Successor) // connect to the successor
	result, err := client.Notify(context.Background(), c.Node.GrpcNode())
	if err != nil {
		fmt.Printf("Error notifying successo: %s:%d \n", c.Successor.IP, c.Successor.Port)
		return false, err
	}
	// fmt.Printf("Remote node has notified %+v! \n", result)
	return result.Value, err
}

// replaceSuccessor replace successor with next one
func (c *Chord) replaceSuccessor(newSuccessor *Node) {
	// replace successor with the next node in successor list
	c.Successor = c.SuccessorList[1]
	c.mutex.Lock()
	c.FingerTable[1] = c.Successor
	c.mutex.Unlock()
}

// Stabilize keep successor and predecessor updated
// Runs periodically
// FIXME if successor failed, replace with next item in successorlist
// ref E.1 - E.3
func (c *Chord) Stabilize() {
	// If node has no successor yet, ignore
	if c.Successor == nil {
		return
	}
	var successorAvailable bool = true
	if !c.Ping(c.Successor) {
		fmt.Printf("Successor failed! %x\n", c.Successor.Identifier)
		successorAvailable = false
		// skip first (start from 1, we already checked 0)
		for i := 1; i < len(c.SuccessorList); i++ {
			if c.Ping(c.SuccessorList[i]) {
				c.replaceSuccessor(c.SuccessorList[i])
				successorAvailable = true
				fmt.Printf("Successor replaced with! %x\n", c.Successor.Identifier)
				break
			}
		}
	}
	// if no successor available, skip stablize
	if !successorAvailable {
		return
	}
	remotePredecessor, successorList := c.GetRemoteStablierData(c.Successor)
	// means successor's predececcor is changed
	if remotePredecessor.Identifier != c.Node.Identifier {
		// if pred(succ) ∈ (n, succ)
		if helpers.BetweenR(remotePredecessor.Identifier, c.Node.Identifier, c.Successor.Identifier) {
			c.mutex.Lock()
			c.FingerTable[1] = remotePredecessor
			c.mutex.Unlock()
			c.Successor = remotePredecessor
			// immediatly update new successor about it's new predecessor
			c.RemoteSuccessorNotify()
		}
	}
	// Update successor list - ref E.3
	c.SuccessorList = successorList
}

// CheckPredecessor keeps predecessor uptodate
// Runs periodically
// ref E.1
func (c *Chord) CheckPredecessor() {
	if c.Predecessor != nil {
		if !c.Ping(c.Predecessor) { // predecessor disconnected !
			c.Predecessor = nil // set nil to be able to update predecessor by notify
		}
	}
}

// FixFingers refreshes finger table entities
// Runs periodically
// ref D - E.1 - finger[k] = (n + 2 ** k-1) Mod M
func (c *Chord) FixFingers() {
	if c.Successor == nil {
		return
	}
	c.FingerFixerIndex++
	if c.FingerFixerIndex > c.m {
		c.FingerFixerIndex = 1
	}

	meint := new(big.Int)
	meint.SetBytes(c.Node.Identifier[:])

	baseint := new(big.Int)
	baseint.SetUint64(2)

	powint := new(big.Int)
	powint.SetInt64(int64(c.FingerFixerIndex - 1))

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
	findSuccessor := c.FindSuccessor(identifier)
	stringidentifier := fmt.Sprintf("%x", identifier)
	c.fingetTableDebug[c.FingerFixerIndex] = stringidentifier
	c.mutex.Lock()
	c.FingerTable[c.FingerFixerIndex] = findSuccessor
	c.mutex.Unlock()
	// first item in fingertable is immediate successor - ref D
	if c.FingerFixerIndex == 1 {
		c.mutex.RLock()
		c.Successor = c.FingerTable[c.FingerFixerIndex]
		c.mutex.RUnlock()
		// immediatly update new successor about it's new predecessor
		c.RemoteSuccessorNotify()
	}
}

// Debug prints successor,predecessor
// Runs periodically
func (c *Chord) Debug() {
	fmt.Printf("Current Node: %s:%d:%x\n", c.Node.IP, c.Node.Port, c.Node.Identifier)
	// if c.Successor != nil {
	// 	fmt.Printf("Current Node Successor: %s:%d:%x\n", c.Successor.IP, c.Successor.Port, c.Successor.Identifier)
	// }
	// if c.Predecessor != nil {
	// 	fmt.Printf("Current Node Predecessor: %s:%d:%x\n", c.Predecessor.IP, c.Predecessor.Port, c.Predecessor.Identifier)
	// }
	// for i := 0; i < len(c.SuccessorList); i++ {
	// 	fmt.Printf("successorList %d: %x\n", i, c.SuccessorList[i].Identifier)
	// }
	for i := 1; i < len(c.FingerTable); i++ {
		fmt.Printf("FingerTable %d: %s, %x\n", i, c.fingetTableDebug[i], c.FingerTable[i].Identifier)
	}
	fmt.Print("\n")
}

// Connect grpc connect to remote node
// FIXME should be cached
func (c *Chord) Connect(remote *Node) pb.ChordClient {
	addr := remote.FullAddr()
	if c.connectionPool[addr] == nil {
		conn, _ := grpc.Dial(addr, grpc.WithInsecure())
		c.connectionPool[addr] = conn
	}
	return pb.NewChordClient(c.connectionPool[addr])
}

// Ping check if remote port is open - using to check predecessor state
// FIXME should be cached
// ref E.1
func (c *Chord) Ping(remote *Node) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", remote.FullAddr(), timeout)
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
