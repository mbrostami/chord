package net

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/mbrostami/chord"
	chordGrpc "github.com/mbrostami/chord/grpc"
	"github.com/mbrostami/chord/helpers"
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

// Fetch retreive data from remote node
func (rs *RemoteNodeSenderGrpc) Fetch(remoteNode *chord.RemoteNode, key [helpers.HashSize]byte) []byte {
	client := rs.connect(remoteNode) // connect to the successor
	lookup := &chordGrpc.Lookup{
		Key: key[:],
	}
	result, _ := client.Fetch(context.Background(), lookup)
	return result.Data
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

func (rs *RemoteNodeSenderGrpc) GlobalMaintenance(remoteNode *chord.RemoteNode, data []byte) ([]byte, error) {
	// prepare kind of timeout to replace disconnected nodes
	client := rs.connect(remoteNode)

	replicationRequest := &chordGrpc.Replication{
		Data: data,
	}
	replicationResponse, err := client.GlobalMaintenance(context.Background(), replicationRequest)
	if err != nil {
		log.Errorf("Remote GlobalMaintenance failed: %+v \n", err)
		return nil, err
	}
	return replicationResponse.Data, nil
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
