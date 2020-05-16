package chord

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	lgrpc "github.com/mbrostami/chord/grpc"
	"github.com/mbrostami/chord/interfaces"
	"google.golang.org/grpc"
)

type RemoteSender struct {
	connectionPool map[string]*grpc.ClientConn
	remoteNode 	*RemoteNode
	mutex          sync.RWMutex
}

func NewRemoteSender() *interfaces.RemoteSenderInterface {
	var remoteSender interfaces.RemoteSenderInterface
	remoteSender = &RemoteSender{
		connectionPool: make(map[string]*grpc.ClientConn),
	}
	return &remoteSender
}

// FindSuccessor find closest node to the given key in remote node
// ref D
func (rs *RemoteSender) FindSuccessor(remoteNode *RemoteNode, identifier []byte) (*Node, error) {
	client := rs.connect(remoteNode)
	successor, err := client.FindSuccessor(context.Background(), &lgrpc.Lookup{Key: identifier})
	if err != nil {
		// fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
		return nil, err
	}
	return rs.convertToChordNode(successor), err
}

// GetStablizerData successor's (successor list and predecessor)
// to prevent duplicate rpc call, we get both together
// ref E.3
func (rs *RemoteSender) GetStablizerData(remoteNode *RemoteNode, caller *Node) (*Node, *SuccessorList, error) {
	// prepare kind of timeout to replace disconnected nodes
	client := rs.connect(remoteNode)

	stablizerData, err := client.GetStablizerData(context.Background(), rs.convertToGrpcNode(caller))
	if err != nil {
		fmt.Printf("Remote GetStablizerData failed: %+v \n", err)
		return nil, nil, err
	}
	predecessor := rs.convertToChordNode(stablizerData.Predecessor)
	// map grpc successor list to chord.successor list
	successorList := rs.convertToChordSuccessorList(stablizerData.SuccessorList)
	return predecessor, successorList, nil
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (rs *RemoteSender) Notify(remoteNode *RemoteNode, caller *Node) error {
	client := rs.connect(remoteNode) // connect to the successor
	result, err := client.Notify(context.Background(), rs.convertToGrpcNode(caller))
	if err != nil {
		fmt.Printf("Error notifying successo: %s \n", rs.RemoteNode.GetFullAddress())
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
func (rs *RemoteSender) Ping(remoteNode *RemoteNode) bool {
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
func (rs *RemoteSender) connect(remoteNode *RemoteNode) lgrpc.ChordClient {
	addr := remoteNode.GetFullAddress()
	if rs.connectionPool[addr] == nil {
		conn, _ := grpc.Dial(addr, grpc.WithInsecure())
		rs.mutex.Lock()
		defer rs.mutex.Unlock()
		rs.connectionPool[addr] = conn
	}
	return lgrpc.NewChordClient(rs.connectionPool[addr])
}

func (rs *RemoteSender) convertToGrpcNode(node *Node) *lgrpc.Node {
	grpcNode := &lgrpc.Node{
		IP:   node.IP,
		Port: int32(node.Port),
	}
	return grpcNode
}

func (rs *RemoteSender) convertToChordNode(node *lgrpc.Node) *Node {
	return NewNode(node.IP, uint(node.Port))
}

// convertToGrpcSuccessorList make grpc node entity from chord node
func (rs *RemoteSender) convertToGrpcSuccessorList(slist *SuccessorList) []*lgrpc.Node {
	nodes := []*lgrpc.Node{}
	if slist != nil {
		for i := 0; i < len(slist.Nodes); i++ { // keep sorts
			nodes = append(nodes, rs.convertToGrpcNode(slist.Nodes[i]))
		}
	}
	return nodes
}

// convertToChordSuccessorList make grpc node entity from chord node
func (rs *RemoteSender) convertToChordSuccessorList(nlist []*lgrpc.Node) *SuccessorList {
	nodes := NewSuccessorList()
	for i := 0; i < len(nlist); i++ { // keep sorted
		nodes.Nodes[i] = NewNode(nlist[i].IP, uint(nlist[i].Port))
	}
	return nodes
}
