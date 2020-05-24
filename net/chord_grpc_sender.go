package net

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/mbrostami/chord"
	chordGrpc "github.com/mbrostami/chord/grpc"
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type RemoteNodeSenderGrpc struct {
	connectionPool *cache.Cache
}

func NewRemoteNodeSenderGrpc() chord.RemoteNodeSenderInterface {
	var remoteSender chord.RemoteNodeSenderInterface
	remoteSender = &RemoteNodeSenderGrpc{
		connectionPool: cache.New(10*time.Second, 1*time.Minute),
	}
	return remoteSender
}

// FindSuccessor find closest node to the given key in remote node
// ref D
func (rs *RemoteNodeSenderGrpc) FindSuccessor(remoteNode *chord.RemoteNode, identifier [helpers.HashSize]byte) (*chord.Node, error) {
	client := rs.connect(remoteNode)
	successor, err := client.FindSuccessor(context.Background(), &chordGrpc.Lookup{Key: identifier[:]})
	if err != nil {
		log.Errorf("There is no predecessor from: %s:%d - %v - %v\n", remoteNode.IP, remoteNode.Port, successor, err)
		return nil, err
	}
	return chordGrpc.ConvertToChordNode(successor), err
}

// GetStablizerData successor's (successor list and predecessor)
// to prevent duplicate rpc call, we get both together
// ref E.3
func (rs *RemoteNodeSenderGrpc) GetStablizerData(remoteNode *chord.RemoteNode, localNode *chord.Node) (*chord.Node, *chord.SuccessorList, error) {
	// prepare kind of timeout to replace disconnected nodes
	client := rs.connect(remoteNode)

	stablizerData, err := client.GetStablizerData(context.Background(), chordGrpc.ConvertToGrpcNode(localNode))
	if err != nil {
		log.Errorf("Remote GetStablizerData failed: %+v \n", err)
		return nil, nil, err
	}
	predecessor := chordGrpc.ConvertToChordNode(stablizerData.Predecessor)
	// map grpc successor list to chord.successor list
	successorList := chordGrpc.ConvertToChordSuccessorList(stablizerData.SuccessorList, rs)
	return predecessor, successorList, nil
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (rs *RemoteNodeSenderGrpc) Notify(remoteNode *chord.RemoteNode, localNode *chord.Node) error {
	client := rs.connect(remoteNode) // connect to the successor
	result, err := client.Notify(context.Background(), chordGrpc.ConvertToGrpcNode(localNode))
	if err != nil {
		log.Errorf("Error notifying successor: %s err: %v \n", remoteNode.GetFullAddress(), err)
		return err
	}
	if !result.Value {
		log.Error("notify failed")
		return errors.New("notify failed")
	}
	return nil
}

// Store store data in remote node
func (rs *RemoteNodeSenderGrpc) Store(remoteNode *chord.RemoteNode, data []byte) bool {
	client := rs.connect(remoteNode) // connect to the successor
	content := &chordGrpc.Content{
		Data: data,
	}
	result, _ := client.Store(context.Background(), content)
	return result.Value
}

// GetPredecessorList predecessor's (predecessor list)
func (rs *RemoteNodeSenderGrpc) GetPredecessorList(remoteNode *chord.RemoteNode, localNode *chord.Node) (*chord.PredecessorList, error) {
	// prepare kind of timeout to replace disconnected nodes
	client := rs.connect(remoteNode)

	nodeList, err := client.GetPredecessorList(context.Background(), chordGrpc.ConvertToGrpcNode(localNode))
	if err != nil {
		log.Errorf("Remote GetStablizerData failed: %+v \n", err)
		return nil, err
	}
	// map grpc nodes to chord.predecessor list
	predecessorList := chordGrpc.ConvertToChordPredecessorList(nodeList.Nodes, rs)
	return predecessorList, nil
}

// ForwardSync sync data with remote node
func (rs *RemoteNodeSenderGrpc) ForwardSync(remoteNode *chord.RemoteNode, plHash [helpers.HashSize]byte, data []byte, tree *merkle.MerkleTree) (*merkle.MerkleTree, error) {
	// prepare kind of timeout to replace disconnected nodes
	client := rs.connect(remoteNode)

	forwardSync := &chordGrpc.ForwardSyncData{
		Data:                data,
		PredecessorListHash: plHash[:],
	}
	grpcMerkleTree := &chordGrpc.MerkleTree{
		RootHash: tree.MerkleRoot(),
	}
	for _, node := range tree.GetSerializedTree() {
		grpcMerkleTree.Nodes = append(grpcMerkleTree.Nodes, &chordGrpc.MerkleNode{
			Hash:  node.Hash,
			Right: node.Right,
			Left:  node.Left,
		})
	}
	forwardSync.MerkleTree = grpcMerkleTree
	receivedForwardSync, err := client.ForwardSync(context.Background(), forwardSync)
	if err != nil {
		log.Errorf("ForwardSync failed: %+v \n", err)
		return nil, err
	}
	if receivedForwardSync.MerkleTree == nil {
		// means stored in remote node
		return nil, nil
	}
	log.Debugf("remote response tree %+v", receivedForwardSync.MerkleTree.Nodes)
	// find missing items and store on remote node
	return nil, errors.New("Remote node has not the same tree")
}

// Ping check if remote port is open - using to check predecessor state
// FIXME should be cached
// ref E.1
func (rs *RemoteNodeSenderGrpc) Ping(remoteNode *chord.RemoteNode) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", remoteNode.GetFullAddress(), timeout)
	if err != nil {
		log.Errorf("Ping %s:%d error:%v", remoteNode.IP, remoteNode.Port, err)
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

// Connect grpc connect to remote node
func (rs *RemoteNodeSenderGrpc) connect(remoteNode *chord.RemoteNode) chordGrpc.ChordClient {
	addr := remoteNode.GetFullAddress()
	var conn *grpc.ClientConn
	if x, found := rs.connectionPool.Get(addr); found {
		conn = x.(*grpc.ClientConn)
	} else {
		conn, _ = grpc.Dial(addr, grpc.WithInsecure())
		rs.connectionPool.Set(addr, conn, cache.DefaultExpiration)
	}
	return chordGrpc.NewChordClient(conn)
}
