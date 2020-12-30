package chord

import (
	"encoding/json"
	"sort"

	"github.com/mbrostami/chord/helpers"
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
		dstore:          NewDStore(localNode.GetFullAddress()),
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

	// update predecessor list
	// TODO can be replaces ping predecessor
	predecessor, predecessorList, err := r.stabilizer.StartPredecessorList(r.predecessor, r.localNode)
	r.predecessorList.UpdatePredecessorList(r.successor, predecessor, r.localNode, predecessorList)
	if r.predecessor == nil {
		return
	}
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

// SyncData sync local data with successor
// FIXME too many open files
func (r *Ring) SyncData() error {
	// ignore self sync
	if r.successor.Identifier == r.localNode.Identifier {
		return nil
	}
	replicas := 2
	lastPredIndex := replicas - 2

	// in order to sync data with successor, we should know about predecessors first
	if r.predecessorList.Nodes[lastPredIndex] == nil {
		log.Debug("ring:SyncData predecessors are not enough")
		return nil
	}
	// log.Debug("ring:SyncData strated")

	// Make range of data that needs to be synced based on predecessors
	ranges := make(map[int][helpers.HashSize]byte)
	ranges[0] = r.localNode.Identifier
	lastIndex := 0
	for i := 0; i <= lastPredIndex; i++ {
		if r.predecessorList.Nodes[i].Ping() {
			lastIndex++
			ranges[lastIndex] = r.predecessorList.Nodes[i].Identifier
		} else {
			lastPredIndex++
			log.Error("ring:SyncData predecessor ping timeout")
			continue
		}
	}
	// if replica is 2, we only need first predecessor to current node range of data
	localData, rootHash := r.dstore.GetRangeCircular(ranges[lastIndex], ranges[0])
	log.Infof("ring:SyncData db range got %d from %x to %x", len(localData), ranges[lastIndex], ranges[0])
	if localData == nil {
		log.Infof("ring:SyncData there is no data to replicate")
		return nil
	}
	// log.Infof("ring:SyncData root hash: %x", rootHash)
	d := NewData(nil, ranges, rootHash)
	jsonRequest, err := SerializeData(d)

	log.Infof("ring:SyncData sending: %v, %v", d, err)
	jsonResponse, err := r.successor.GlobalMaintenance(jsonRequest)

	if err != nil {
		log.Errorf("ring:SyncData error in remote global maintenance: %v", err)
		return err
	}
	if jsonResponse == nil {
		log.Info("ring:SyncData data is already synced!")
		return nil
	}

	responseData := UnserializeData(jsonResponse)

	// store missing data in remote node
	for id, record := range localData {
		if responseData.Records[id] == nil {
			// log.Infof("ring:SyncData store on remote node: %x", responseData.Records[id])
			r.successor.Store(record.Content)
		}
	}

	// store missing data in local node
	for id, record := range responseData.Records {
		if localData[id] == nil {
			// log.Infof("ring:SyncData store on local node: %x", localData[id])
			r.Store(record.Content)
		}
	}
	// log.Infof("ring:SyncData different rows: %+v", rows)
	return nil
}

// GlobalMaintenance gets data information from predecessor to sync missing data
func (r *Ring) GlobalMaintenance(jsonData []byte) ([]byte, error) {
	replicas := 2
	lastPredIndex := replicas - 1
	data := UnserializeData(jsonData)
	log.Debugf("ring:GlobalMaintenance strated %v", string(jsonData))

	ranges := data.Ranges // use received ranges
	// if replica is 2, we only need second predecessor to predecessor range of data
	allKeysInRange, rootHash := r.dstore.GetRangeCircular(ranges[lastPredIndex], ranges[0])
	// log.Infof("ring:GlobalMaintenance db range got %d from %x to %x", len(allKeysInRange), ranges[lastPredIndex], ranges[0])
	if allKeysInRange == nil {
		// log.Infof("ring:GlobalMaintenance there is no data to make hash")
		return nil, nil
	}
	// if new root hash is the same as roothash in json data, means data is already synced
	if helpers.Equal(rootHash, data.RootHash) {
		return nil, nil
	}

	newData := NewData(allKeysInRange, ranges, rootHash)
	return SerializeData(newData)
}

func (r *Ring) Fetch(key [helpers.HashSize]byte) []byte {
	return r.dstore.Get(key)
}

// Store store data
// @todo replicate to the successor (required replications)
// ref E.3
func (r *Ring) Store(jsonData []byte) bool {
	record := &Record{}
	json.Unmarshal(jsonData, &record)
	log.Warnf("ring:store put %x", record.Hash())
	return r.dstore.PutRecord(*record)
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
	// log.Debugf("Current Node: %s:%d:%x", r.localNode.IP, r.localNode.Port, r.localNode.Identifier)
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
	records := r.dstore.GetAll()
	// sort data
	sortedKeys := make([]string, 0, len(records))
	for k := range records {
		sortedKeys = append(sortedKeys, string(k[:]))
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		var key [helpers.HashSize]byte
		copy(key[:helpers.HashSize], []byte(k)[:helpers.HashSize])
		log.Debugf("db: %s : %x", records[key].Content, records[key].Identifier)
	}
	log.Debugf("\n")
}
