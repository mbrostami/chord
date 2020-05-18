package net

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mbrostami/chord"
	chordGrpc "github.com/mbrostami/chord/grpc"
	"github.com/mbrostami/chord/helpers"
	"google.golang.org/grpc"
)

type RemoteNodeSenderGrpc struct {
	connectionPool map[string]*grpc.ClientConn
	mutex          sync.RWMutex
}

func NewRemoteNodeSenderGrpc() chord.RemoteNodeSenderInterface {
	var remoteSender chord.RemoteNodeSenderInterface
	remoteSender = &RemoteNodeSenderGrpc{
		connectionPool: make(map[string]*grpc.ClientConn),
	}
	return remoteSender
}

// FindSuccessor find closest node to the given key in remote node
// ref D
func (rs *RemoteNodeSenderGrpc) FindSuccessor(remoteNode *chord.RemoteNode, identifier [helpers.HashSize]byte) (*chord.Node, error) {
	client := rs.connect(remoteNode)
	successor, err := client.FindSuccessor(context.Background(), &chordGrpc.Lookup{Key: identifier[:]})
	if err != nil {
		// fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
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
		fmt.Printf("Remote GetStablizerData failed: %+v \n", err)
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
		fmt.Printf("Error notifying successor: %s err: %v \n", remoteNode.GetFullAddress(), err)
		return err
	}
	if !result.Value {
		return errors.New("notify failed")
	}
	// fmt.Printf("Remote node has notified %+v! \n", result)
	return nil
}

// Ping check if remote port is open - using to check predecessor state
// FIXME should be cached
// ref E.1
func (rs *RemoteNodeSenderGrpc) Ping(remoteNode *chord.RemoteNode) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", remoteNode.GetFullAddress(), timeout)
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

// Connect grpc connect to remote node
func (rs *RemoteNodeSenderGrpc) connect(remoteNode *chord.RemoteNode) chordGrpc.ChordClient {
	addr := remoteNode.GetFullAddress()
	if rs.connectionPool[addr] == nil {
		conn, _ := grpc.Dial(addr, grpc.WithInsecure())
		rs.mutex.Lock()
		defer rs.mutex.Unlock()
		rs.connectionPool[addr] = conn
	}
	return chordGrpc.NewChordClient(rs.connectionPool[addr])
}
