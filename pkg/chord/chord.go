package chord

// Based on https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"sync"

	"github.com/mbrostami/chord/pkg/helpers"
)

// MSIZE is the number of bits in identifier
// in fact only O(log n) are distinct
// ref D - Theorem IV.2
const MSIZE int = sha256.Size * 8

// RSIZE is the number of records in successor list
// could be (log n) refer to Theorem IV.5
// Increasing r makes the system more robust
// ref E.3 - Theorem IV.5
const RSIZE int = sha256.Size // 32

// Chord chord ring
// ref B
type Chord struct {
	clientAdapter    ClientInterface // client adapter
	Successor        *Node           // next node
	Predecessor      *Node           // previous node
	SuccessorList    *SuccessorList
	FingerTable      map[int]*Node // ref D
	FingerFixerIndex int           // to use in fixFinger
	mutex            sync.RWMutex
	m                int // ref D - MSIZE
	r                int // ref E.3 - RSIZE
	Node             *Node
	fingetTableDebug map[int]string // FIXME remove this one, it's for debug
}

// NewRing create new chord ring
// ref E.1
func NewRing(ip string, port int, clientAdapter ClientInterface) *Chord {
	chord := &Chord{
		clientAdapter:    clientAdapter,
		SuccessorList:    NewSuccessorList(),
		FingerTable:      make(map[int]*Node),
		FingerFixerIndex: 0,
		m:                MSIZE,
		r:                RSIZE,
		fingetTableDebug: make(map[int]string),
	}

	chord.Node = NewNode(ip, port)
	chord.Successor = chord.Node
	chord.Predecessor = nil
	return chord
}

// Join throw first node
// ref E.1
func (c *Chord) Join(remote *Node) error {
	successor, err := c.clientAdapter.FindSuccessor(remote, c.Node.Identifier[:])
	if err != nil {
		fmt.Printf("Error Join: %v", err)
		return err
	}
	//fmt.Printf("Join: got successor %s:%d! \n", successor.IP, successor.Port)
	c.Predecessor = nil
	c.Successor = NewNode(successor.IP, int(successor.Port))

	c.mutex.Lock()
	c.FingerTable[1] = c.Successor
	c.mutex.Unlock()
	c.clientAdapter.Notify(c.Successor, c.Node)
	return nil
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (c *Chord) Notify(node *Node) bool {
	// (c.predecessor is nil or node ∈ (c.predecessor, n))
	if c.Predecessor == nil {
		c.Predecessor = node
		// First node in the network has itself as successor
		if c.Successor.Identifier == c.Node.Identifier {
			c.Successor = node
			c.clientAdapter.Notify(c.Successor, c.Node) // notify successor
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
		return c.Node
	}
	// id ∈ (n, successor]
	if helpers.BetweenR(identifier, c.Node.Identifier, c.Successor.Identifier) {
		return c.Successor
	}
	nextNode := c.closestPrecedingNode(identifier)
	if nextNode.Identifier == c.Node.Identifier { // current node is the only node in figer table
		return nextNode
	}
	nextNodeSuccessor, err := c.clientAdapter.FindSuccessor(nextNode, identifier[:])
	if err != nil { // unexpected error on successor
		fmt.Printf("Unexpected error from successor %v", err)
		return nil
	}
	return nextNodeSuccessor
}

// TODO E.3
// ref D
func (c *Chord) closestPrecedingNode(identifier [sha256.Size]byte) *Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	var closesNodeInFingerTable *Node
	for m := len(c.FingerTable); m > 0; m-- {
		if c.FingerTable[m] != nil {
			// finger[i] ∈ (n, id)
			if helpers.Between(c.FingerTable[m].Identifier, c.Node.Identifier, identifier) {
				closesNodeInFingerTable = c.FingerTable[m]
				break
			}
		}
	}

	// ref E.3 - modified version of closestPrecedingNode will also check the successor list
	// to find closest node to the identifier
	for i := 0; i < len(c.SuccessorList.Nodes); i++ {
		// successorList[i] ∈ (fingerTable[k], id)
		// if successor item is closer to identifier than the found fingertable item, then return successor item
		if closesNodeInFingerTable != nil {
			if helpers.BetweenR(c.SuccessorList.Nodes[i].Identifier, closesNodeInFingerTable.Identifier, identifier) {
				return c.SuccessorList.Nodes[i]
			}
		} else {
			// successorList[i] ∈ (n, id)
			if helpers.BetweenR(c.SuccessorList.Nodes[i].Identifier, c.Node.Identifier, identifier) {
				return c.SuccessorList.Nodes[i]
			}
		}
	}
	if closesNodeInFingerTable != nil {
		return closesNodeInFingerTable
	}

	return c.Node
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
	return c.Node
}

// GetSuccessorList returns unsorted successor list
// ref E.3
func (c *Chord) GetSuccessorList() *SuccessorList {
	return c.SuccessorList
}

// replaceSuccessor replace successor with next one
func (c *Chord) replaceSuccessor(newSuccessor *Node) {
	// replace successor with the next node in successor list
	c.Successor = newSuccessor
	c.mutex.Lock()
	c.FingerTable[1] = c.Successor
	c.mutex.Unlock()
}

// Stabilize keep successor and predecessor updated
// Runs periodically
// FIXME if successor failed, replace with next item in successorlist
// ref E.1 - E.3
func (c *Chord) Stabilize() {
	var successorAvailable bool = true
	var remotePredecessor *Node
	var successorList *SuccessorList
	var err error
	remotePredecessor, successorList, err = c.clientAdapter.GetStablizerData(c.Successor, c.Node)
	if err != nil {
		fmt.Printf("Successor failed! %x\n", c.Successor.Identifier)
		successorAvailable = false
		// skip first (start from 1, we already checked 0)
		for i := 1; i < len(c.SuccessorList.Nodes); i++ {
			remotePredecessor, successorList, err = c.clientAdapter.GetStablizerData(c.SuccessorList.Nodes[i], c.Node)
			if err == nil {
				c.replaceSuccessor(c.SuccessorList.Nodes[i])
				successorAvailable = true
				fmt.Printf("Successor replaced with! %x\n", c.Successor.Identifier)
				break
			}
		}
	}
	// if no successor available, skip stablize
	if !successorAvailable {
		c.replaceSuccessor(c.Node)
		return
	}
	// means successor's predececcor is changed
	if remotePredecessor.Identifier != c.Node.Identifier {
		// if pred(succ) ∈ (n, succ)
		if helpers.BetweenR(remotePredecessor.Identifier, c.Node.Identifier, c.Successor.Identifier) {
			c.mutex.Lock()
			c.FingerTable[1] = remotePredecessor
			c.mutex.Unlock()
			c.Successor = remotePredecessor
			// immediatly update new successor about it's new predecessor
			c.clientAdapter.Notify(c.Successor, c.Node)
		}
	}
	// Update successor list - ref E.3
	c.updateSuccessorList(successorList)
}

// Update successor list - ref E.3
func (c *Chord) updateSuccessorList(successorList *SuccessorList) {
	if successorList == nil || c.Successor == nil {
		return
	}
	slist := NewSuccessorList()
	slist.Nodes[0] = c.Successor // replace first item with successor itself
	index := 1
	// FIXME read lock
	for i := 0; i < len(successorList.Nodes); i++ {
		if len(slist.Nodes) >= c.r { // prevent overloading successorlist (max(r)=(log N)) ref E.3
			break
		}
		chorNode := NewNode(successorList.Nodes[i].IP, int(successorList.Nodes[i].Port))
		// ignore same nodes
		if chorNode.Identifier == c.Node.Identifier {
			continue
		}
		slist.Nodes[index] = chorNode
		index++
	}
	// FIXME write lock
	c.SuccessorList = slist
}

// CheckPredecessor keeps predecessor uptodate
// Runs periodically
// ref E.1
func (c *Chord) CheckPredecessor() {
	if c.Predecessor != nil {
		if !c.clientAdapter.Ping(c.Predecessor) { // predecessor disconnected !
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
	c.fingetTableDebug[c.FingerFixerIndex] = fmt.Sprintf("%x", identifier)
	c.mutex.Lock()
	c.FingerTable[c.FingerFixerIndex] = findSuccessor
	c.mutex.Unlock()
	// first item in fingertable is immediate successor - ref D
	if c.FingerFixerIndex == 1 {
		c.mutex.RLock()
		c.Successor = c.FingerTable[c.FingerFixerIndex]
		c.mutex.RUnlock()
		// immediatly update new successor about it's new predecessor
		c.clientAdapter.Notify(c.Successor, c.Node)
	}
}

// Debug prints successor,predecessor
// Runs periodically
func (c *Chord) Debug() {
	fmt.Printf("Current Node: %s:%d:%x\n", c.Node.IP, c.Node.Port, c.Node.Identifier)
	if c.Successor != nil {
		fmt.Printf("Current Node Successor: %s:%d:%x\n", c.Successor.IP, c.Successor.Port, c.Successor.Identifier)
	}
	// if c.Predecessor != nil {
	// 	fmt.Printf("Current Node Predecessor: %s:%d:%x\n", c.Predecessor.IP, c.Predecessor.Port, c.Predecessor.Identifier)
	// }
	// for i := 0; i < len(c.SuccessorList.Nodes); i++ {
	// fmt.Printf("successorList %d: %x\n", i, c.SuccessorList.Nodes[i].Identifier)
	// }
	// for i := 1; i < len(c.FingerTable); i++ {
	// 	fmt.Printf("FingerTable %d: %s, %x\n", i, c.fingetTableDebug[i], c.FingerTable[i].Identifier)
	// }
	fmt.Print("\n")
}
