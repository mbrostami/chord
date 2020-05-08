package server

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	dstoregrpc "github.com/mbrostami/chord/internal/grpc/dstore"
	"github.com/mbrostami/chord/pkg/chord"
)

// DStoreServer grpc server
type DStoreServer struct {
	dstoregrpc.UnimplementedDStoreServer
	ChordRing *chord.Chord
	database  map[[sha256.Size]byte]string
	// database map()
}

// NewDStoreServer create new storage server
func NewDStoreServer(chord *chord.Chord) *DStoreServer {
	return &DStoreServer{
		database:  make(map[[sha256.Size]byte]string),
		ChordRing: chord,
	}
}

// Store store key value
func (s *DStoreServer) Store(ctx context.Context, keyValue *dstoregrpc.KeyValue) (*dstoregrpc.StoreResponse, error) {
	fmt.Printf("SERVER: Storing data in node key: %x value: %s \n", keyValue.Key, string(keyValue.Value))
	response := &dstoregrpc.StoreResponse{}
	var key [sha256.Size]byte
	copy(key[:sha256.Size], keyValue.Key[:sha256.Size])
	s.database[key] = string(keyValue.Value)
	for i := 0; i < len(s.ChordRing.SuccessorList.Nodes); i++ {
		node := &dstoregrpc.Node{}
		node.IP = s.ChordRing.SuccessorList.Nodes[i].IP
		node.Port = int32(s.ChordRing.SuccessorList.Nodes[i].Port)
		response.Nodes = append(response.Nodes, node)
	}
	return response, nil
}

// Get find the value of the given key
func (s *DStoreServer) Get(ctx context.Context, lookup *dstoregrpc.Lookup) (*dstoregrpc.LookupResponse, error) {
	fmt.Printf("SERVER: Getting data %x \n", lookup.Key)
	response := &dstoregrpc.LookupResponse{}
	var search [sha256.Size]byte
	copy(search[:sha256.Size], lookup.Key[:sha256.Size])
	if value, exist := s.database[search]; exist {
		fmt.Printf("SERVER: Found data %s \n", value)
		response.Value = []byte(value)
		return response, nil
	}
	for i := 0; i < len(s.ChordRing.SuccessorList.Nodes); i++ {
		node := &dstoregrpc.Node{}
		node.IP = s.ChordRing.SuccessorList.Nodes[i].IP
		node.Port = int32(s.ChordRing.SuccessorList.Nodes[i].Port)
		fmt.Printf("SERVER: returning successorlist %s:%d \n", node.IP, node.Port)
		response.Nodes = append(response.Nodes, node)
	}
	if response.Nodes == nil {
		return nil, errors.New("key not found")
	}
	return response, nil
}
