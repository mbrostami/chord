package clientadapter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	pb "github.com/mbrostami/chord/internal/grpc"
	"github.com/mbrostami/chord/internal/server"
	"github.com/mbrostami/chord/pkg/chord"
	"google.golang.org/grpc"
)

// GrpcClient inherited chord.ClientInterface
type GrpcClient struct {
	connectionPool map[string]*grpc.ClientConn
	mutex          sync.RWMutex
}

// NewClient make grpc client adapter
func NewClient() *GrpcClient {
	chord := &GrpcClient{
		connectionPool: make(map[string]*grpc.ClientConn),
	}
	return chord
}

// FindSuccessor find closest node to the given key in remote node
// ref D
func (c *GrpcClient) FindSuccessor(remote *chord.Node, identifier []byte) (*chord.Node, error) {
	client := c.connect(remote)
	successor, err := client.FindSuccessor(context.Background(), &pb.Lookup{Key: identifier})
	if err != nil {
		// fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
		return nil, err
	}
	chordNode := chord.NewNode(successor.IP, int(successor.Port))
	return chordNode, err
}

// GetStablizerData successor's (successor list and predecessor)
// to prevent duplicate rpc call, we get both together
// ref E.3
func (c *GrpcClient) GetStablizerData(remote *chord.Node, node *chord.Node) (*chord.Node, *chord.SuccessorList, error) {
	// prepare kind of timeout to replace disconnected nodes
	client := c.connect(remote)

	stablizerData, err := client.GetStablizerData(context.Background(), server.ConvertToGrpcNode(node))
	if err != nil {
		fmt.Printf("Remote GetStablizerData failed: %+v \n", err)
		return nil, nil, err
	}
	predecessor := server.ConvertToChordNode(stablizerData.Predecessor)
	// map grpc successor list to chord.successor list
	successorList := server.ConvertToChordSuccessorList(stablizerData.SuccessorList)
	return predecessor, successorList, nil
}

// Notify update predecessor
// is being called periodically by predecessor or new node
// ref E.1
func (c *GrpcClient) Notify(remote *chord.Node, node *chord.Node) error {
	client := c.connect(remote) // connect to the successor
	result, err := client.Notify(context.Background(), server.ConvertToGrpcNode(node))
	if err != nil {
		fmt.Printf("Error notifying successo: %s:%d \n", remote.IP, remote.Port)
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
func (c *GrpcClient) Ping(remote *chord.Node) bool {
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

// Connect grpc connect to remote node
func (c *GrpcClient) connect(remote *chord.Node) pb.ChordClient {
	addr := remote.FullAddr()
	if c.connectionPool[addr] == nil {
		conn, _ := grpc.Dial(addr, grpc.WithInsecure())
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.connectionPool[addr] = conn
	}
	return pb.NewChordClient(c.connectionPool[addr])
}
