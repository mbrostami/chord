package chord

import (
	"errors"

	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
	log "github.com/sirupsen/logrus"
)

type Ring struct {
	localNode       *Node
	remoteSender    RemoteNodeSenderInterface
	fingerTable     *FingerTable
	successorList   *SuccessorList
	predecessorList *PredecessorList
	stabilizer      *Stabilizer
	predecessor     *RemoteNode
	successor       *RemoteNode
	dstore          *DStore
}

func NewRing(localNode *Node, remoteSender RemoteNodeSenderInterface) RingInterface {
	var ring RingInterface
	successorList := NewSuccessorList()
	predecessorList := NewPredecessorList()
	ring = &Ring{
		localNode:       localNode,
		remoteSender:    remoteSender,
		fingerTable:     NewFingerTable(),
		stabilizer:      NewStabilizer(successorList, predecessorList),
		successorList:   successorList,
		predecessorList: predecessorList,
		successor:       NewRemoteNode(localNode, remoteSender),
		predecessor:     nil,
		dstore:          NewDStore(),
	}
	return ring
}

// Join throw first node
// ref E.1
func (r *Ring) Join(remoteNode *RemoteNode) error {
	successor, err := remoteNode.FindSuccessor(r.localNode.Identifier)
	if err != nil {
		log.Errorf("Error Join: %v", err)
		return err
	}
	//fmt.Printf("Join: got successor %s:%d! \n", successor.IP, successor.Port)
	r.predecessor = nil
	r.successor = successor
	r.fingerTable.Set(1, r.successor)
	r.successor.Notify(r.localNode)
	return nil
}

func (r *Ring) GetLocalNode() *Node {
	return r.localNode
}

func (r *Ring) FindSuccessor(identifier [helpers.HashSize]byte) *RemoteNode {
	// fmt.Printf("FindSuccessor: start looking for key %x \n", identifier)
	if r.successor.Identifier == r.localNode.Identifier {
		return NewRemoteNode(r.localNode, r.remoteSender)
	}
	// id ∈ (n, successor]
	if helpers.BetweenR(identifier, r.localNode.Identifier, r.successor.Identifier) {
		return r.successor
	}
	closestRemoteNode := r.fingerTable.ClosestPrecedingNode(identifier, r.localNode)
	successorListClosestNode := r.successorList.ClosestPrecedingNode(identifier, r.localNode, closestRemoteNode)
	if successorListClosestNode != nil {
		closestRemoteNode = successorListClosestNode
	} else if closestRemoteNode == nil {
		closestRemoteNode = NewRemoteNode(r.localNode, r.remoteSender) // make a copy local node as remote node
	}
	if closestRemoteNode.Identifier == r.localNode.Identifier { // current node is the only node in figer table
		return closestRemoteNode // return local node
	}
	nextNodeSuccessor, err := closestRemoteNode.FindSuccessor(identifier)
	if err != nil { // unexpected error on successor
		log.Errorf("Unexpected error from successor %v", err)
		return nil
	}
	return nextNodeSuccessor
}

// Stabilize keep successor and predecessor updated
// Runs periodically
// ref E.1 - E.3
func (r *Ring) Stabilize() {
	successor, successorList, err := r.stabilizer.StartSuccessorList(r.successor, r.localNode)
	if err != nil {
		// all successors are failed
		r.successor = NewRemoteNode(r.localNode, r.remoteSender)
		return
	}
	// Update successor list - ref E.3
	r.successorList.UpdateSuccessorList(successor, r.predecessor, r.localNode, successorList)
	// If successor is changed while stabilizing
	if successor.Identifier != r.successor.Identifier {
		r.successor = successor
		r.fingerTable.Set(1, r.successor)
		// immediatly update new successor about it's new predecessor
		r.successor.Notify(r.localNode)
	}

	if r.predecessor == nil {
		return
	}
	// update predecessor list
	// TODO can be replaces ping predecessor
	predecessor, predecessorList, err := r.stabilizer.StartPredecessorList(r.predecessor, r.localNode)
	r.predecessorList.UpdatePredecessorList(r.successor, predecessor, r.localNode, predecessorList)
	// If successor is changed while stabilizing
	if predecessor.Identifier != r.predecessor.Identifier {
		r.predecessor = predecessor
	}
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (r *Ring) Notify(caller *Node) bool {
	// (c.predecessor is nil or node ∈ (c.predecessor, n))
	if r.predecessor == nil {
		r.predecessor = NewRemoteNode(caller, r.remoteSender)
		// First node in the network has itself as successor
		log.Info("Notify applying!")
		if r.successor.Identifier == r.localNode.Identifier {
			log.Info("Bootstrap successor is changed!")
			r.successor = r.predecessor
			r.successor.Notify(r.localNode)
		}
		return true
	}
	if helpers.Between(caller.Identifier, r.predecessor.Identifier, r.localNode.Identifier) {
		r.predecessor = NewRemoteNode(caller, r.remoteSender)
		return true
	}
	return false
}

func (r *Ring) CheckPredecessor() {
	if r.predecessor != nil {
		if !r.predecessor.Ping() {
			r.predecessor = nil // set nil to be able to update predecessor by notify
		}
	}
}

// FixFingers refreshes finger table entities
// Runs periodically
// ref D - E.1 - finger[k] = (n + 2 ** k-1) Mod M
func (r *Ring) FixFingers() {
	if r.successor == nil {
		return
	}
	index, identifier := r.fingerTable.CalculateIdentifier(r.localNode)
	remoteNode := r.FindSuccessor(identifier)
	r.fingerTable.Set(index, remoteNode)
	if index == 1 && remoteNode.Identifier != r.successor.Identifier { // means it's first entry of fingerTable (first entry should be always the next successor of current node)
		r.successor = remoteNode
		// immediatly update new successor about it's new predecessor
		r.successor.Notify(r.localNode)
	}
}

// GetSuccessorList returns unsorted successor list
// ref E.3
func (r *Ring) GetSuccessorList() *SuccessorList {
	return r.successorList
}

// GetPredecessorList returns unsorted predecessor list
func (r *Ring) GetPredecessorList(caller *Node) *PredecessorList {
	return r.predecessorList
}

// GetStabilizerData return predecessor and successor list
func (r *Ring) GetStabilizerData(caller *Node) (*RemoteNode, *SuccessorList) {
	remoteCaller := NewRemoteNode(caller, r.remoteSender)
	predecessor := r.GetPredecessor(remoteCaller)
	return predecessor, r.GetSuccessorList()
}

// Store store data + make merkle tree
// ref E.3
func (r *Ring) Store(data []byte) bool {
	log.Debug("ring:store start")
	if r.predecessor == nil {
		log.Debug("predecessor is nil")
		return false
	}
	hash := helpers.Hash(string(data))
	// check if hash ∈ (c.predecessor, n]
	if !helpers.BetweenR(hash, r.predecessor.Identifier, r.localNode.Identifier) {
		log.Debugf("data hash is not between %x and %x , hash: %x", r.predecessor.Identifier, r.localNode.Identifier, hash)
		// reject storing values which current node is not responsible for
		return false
	}
	if r.predecessorList.Nodes[DBREPLICAS] == nil {
		log.Debugf("ring:Store predecessor list is not updated %d : %d", DBREPLICAS, len(r.predecessorList.Nodes))
		// not possible til predecessor list updated
		return false
	}
	// all data in this range ∈ (r.predecessorList[DBREPLICAS], r.localNode.Identifier]
	allKeysInRange := r.dstore.GetRange(r.predecessorList.Nodes[DBREPLICAS].Identifier, r.localNode.Identifier)

	if len(allKeysInRange) == 0 {
		log.Debugf("Current node: %x", r.localNode.Identifier)
		return r.dstore.Put(hash, data)
	}
	var list []merkle.Content
	//Build list of Content to build tree
	for _, value := range allKeysInRange {
		list = append(list, merkle.TestContent{X: *value})
	}
	//Create a new Merkle Tree from the list of Content
	tree, err := merkle.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}

	// Make a hash out of predecessorList to check with replica node, if it has the same predecessorlist and name it RHash
	var plhash [helpers.HashSize]byte
	// add current node to hash, because successor has already this node in predecessor list
	// so successor can check normally by hasing(DBREPLICAS + 1 predecessors)
	plhash = helpers.Hash(string(plhash[:]) + string(r.localNode.Identifier[:]))
	for i := 0; i < DBREPLICAS; i++ {
		plhash = helpers.Hash(string(plhash[:]) + string(r.predecessorList.Nodes[i].Identifier[:]))
	}
	serializedTreeNodes, err := r.successor.ForwardSync(plhash, data, tree)
	if err != nil {
		log.Debugf("ring: forward sync faild: %x", r.successor.Identifier)
		return false
	}

	if serializedTreeNodes == nil {
		return r.dstore.Put(hash, data)
	}
	// Send merkleTree + new data + RHash
	// receive response and store data locally , then return true.
	log.Debugf("ring: response merkle tree: %+v", serializedTreeNodes)
	// log.Debugf("Current node: %x", r.localNode.Identifier)
	return false
}

// ForwardSync to sync
func (r *Ring) ForwardSync(newData []byte, predecessorListHash [helpers.HashSize]byte, serializedData []*merkle.SerializedNode) ([]*merkle.SerializedNode, error) {
	log.Debug("ring:ForwardSync start")
	if r.predecessor == nil {
		log.Debug("predecessor is nil")
		return nil, errors.New("predecessor is nil")
	}
	if r.predecessorList.Nodes[DBREPLICAS+1] == nil {
		log.Debugf("ring:ForwardSync predecessor list is not completed %d : %d", DBREPLICAS+1, len(r.predecessorList.Nodes))
		if len(r.predecessorList.Nodes) > DBREPLICAS+1 {
			for i := 0; i < len(r.predecessorList.Nodes); i++ {
				log.Debugf("ring:ForwardSync predecessor list i: %d: %x", i, r.predecessorList.Nodes[i].Identifier)
			}
		}
		// not possible til predecessor list updated
		return nil, errors.New("predecessor list is not updated")
	}

	// calculate predecessors list hash
	var plhash [helpers.HashSize]byte
	for i := 0; i < DBREPLICAS+1; i++ {
		plhash = helpers.Hash(string(plhash[:]) + string(r.predecessorList.Nodes[i].Identifier[:]))
	}

	if !helpers.Equal(plhash, predecessorListHash) {
		return nil, errors.New("ring:ForwardSync predecessor lists are not same")
	}

	hash := helpers.Hash(string(newData))
	// check if hash ∈ (c.predecessor[DBREPLICAS+1], c.predecessor] because it's coming from predecessor
	if !helpers.BetweenR(hash, r.predecessorList.Nodes[DBREPLICAS+1].Identifier, r.predecessor.Identifier) {
		log.Debugf("ring: ForwardSync data hash is not between %x and %x , hash: %x", r.predecessor.Identifier, r.localNode.Identifier, hash)
		// reject storing values which current node is not responsible for
		return nil, errors.New("hash is not between")
	}

	// all data in this range ∈ (r.predecessorList[DBREPLICAS + 1], r.predecessor.Identifier]
	allKeysInRange := r.dstore.GetRange(r.predecessorList.Nodes[DBREPLICAS+1].Identifier, r.predecessor.Identifier)

	if len(allKeysInRange) == 0 {
		log.Debugf("ring: ForwardSync Current node: %x", r.localNode.Identifier)
		r.dstore.Put(hash, newData)
		return nil, nil
	}
	var list []merkle.Content
	//Build list of Content to build tree
	for _, value := range allKeysInRange {
		list = append(list, merkle.TestContent{X: *value})
	}
	//Create a new Merkle Tree from the list of Content
	tree, err := merkle.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}
	diffs, dataToDelete, err := tree.Diffs(serializedData)
	if err != nil {
		log.Errorf("Data should be deleted: %+v", dataToDelete)
		return nil, err
	}
	if dataToDelete != nil {
		log.Debugf("Data should be deleted: %+v", dataToDelete)
	}
	if diffs != nil {
		return diffs, nil
	}
	r.dstore.Put(hash, newData)
	return nil, nil
}

func (r *Ring) GetPredecessor(caller *RemoteNode) *RemoteNode {
	if r.predecessor != nil {
		// extension on chord
		if helpers.Between(caller.Identifier, r.predecessor.Identifier, r.localNode.Identifier) {
			r.predecessor = caller // update predecessor if caller is closer than predecessor
		}
		return r.predecessor
	}
	return NewRemoteNode(r.localNode, r.remoteSender)
}

// Verbose prints successor,predecessor
// Runs periodically
func (r *Ring) Verbose() {
	// log.Debugf("Current Node: %s:%d:%x\n", r.localNode.IP, r.localNode.Port, r.localNode.Identifier)
	// if r.successor != nil {
	// 	log.Debugf("Current Node Successor: %s:%d:%x\n", r.successor.IP, r.successor.Port, r.successor.Identifier)
	// }
	// if r.predecessor != nil {
	// log.Debugf("Current Node Predecessor: %s:%d:%x\n", r.predecessor.IP, r.predecessor.Port, r.predecessor.Identifier)
	// }
	// for i := 0; i < len(r.successorList.Nodes); i++ {
	// 	log.Debugf("successorList %d: %x\n", i, r.successorList.Nodes[i].Identifier)
	// }
	// for i := 0; i < len(r.predecessorList.Nodes); i++ {
	// 	log.Debugf("predecessorList %d: %x\n", i, r.predecessorList.Nodes[i].Identifier)
	// }
	// for i := 1; i < len(r.fingerTable.Table); i++ {
	// 	log.Debugf("FingerTable %d: %x\n", i, r.fingerTable.Table[i].Identifier)
	// }
	// log.Debugf("\n")
}
