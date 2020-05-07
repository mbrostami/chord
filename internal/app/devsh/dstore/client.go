package dstore

import (
	"context"
	"fmt"
	"sync"

	dstoregrpc "github.com/mbrostami/chord/internal/grpc/dstore"
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

// Store store data
func (c *GrpcClient) Store(remote *chord.Node, key [chord.IdentifierSize]byte, value []byte) (int, error) {
	client := c.connect(remote)
	kv := &dstoregrpc.KeyValue{Key: key[:], Value: value}
	fmt.Printf("CLIENT: Storing data from node %s, %x \n", remote.FullAddr(), key)
	response, err := client.Store(context.Background(), kv)
	if err != nil {
		// fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
		return 0, err
	}

	numberOfWrites := 1
	// TODO make a worker queue
	for _, node := range response.Nodes {
		chordNode := server.ConvertDstoreNodeToChordNode(node)
		client := c.connect(chordNode)
		_, err := client.Store(context.Background(), kv)
		if err == nil {
			numberOfWrites++
		}
	}
	return numberOfWrites, nil
}

// Get store data
func (c *GrpcClient) Get(remote *chord.Node, key [chord.IdentifierSize]byte) ([]byte, error) {
	client := c.connect(remote)
	lookup := &dstoregrpc.Lookup{Key: key[:]}
	fmt.Printf("CLIENT: Getting data from node %s, %x \n", remote.FullAddr(), key)
	response, err := client.Get(context.Background(), lookup)
	if err != nil {
		// fmt.Printf("There is no predecessor from: %s:%d - %v - %v\n", remote.IP, remote.Port, successor, err)
		return nil, err
	}
	if len(response.Nodes) > 0 {
		return response.Value, nil
	}
	return nil, nil
}

// Connect grpc connect to remote node
func (c *GrpcClient) connect(remote *chord.Node) dstoregrpc.DStoreClient {
	addr := remote.FullAddr()
	if c.connectionPool[addr] == nil {
		conn, _ := grpc.Dial(addr, grpc.WithInsecure())
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.connectionPool[addr] = conn
	}
	return dstoregrpc.NewDStoreClient(c.connectionPool[addr])
}
