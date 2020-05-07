package server

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	dstoregrpc "github.com/mbrostami/chord/internal/grpc/dstore"
	"github.com/mbrostami/chord/pkg/chord"
	"github.com/mbrostami/chord/pkg/helpers"
)

type Entity struct {
	Identity []byte
	Value    string
}

// DStoreServer grpc server
type DStoreServer struct {
	dstoregrpc.UnimplementedDStoreServer
	ChordRing *chord.Chord
	database  map[int]*Entity
	// database map()
}

// Store store key value
func (s *DStoreServer) Store(ctx context.Context, keyValue *dstoregrpc.KeyValue) (*dstoregrpc.StoreResponse, error) {
	fmt.Printf("SERVER: Storing data in node %x : value: %s \n", keyValue.Key, string(keyValue.Value))
	response := &dstoregrpc.StoreResponse{}
	if s.database == nil {
		s.database = make(map[int]*Entity)
	}
	s.database[len(s.database)] = &Entity{
		Identity: keyValue.Key,
		Value:    string(keyValue.Value),
	}
	for i := 0; i < len(s.ChordRing.SuccessorList.Nodes); i++ {
		node := &dstoregrpc.Node{}
		node.IP = s.ChordRing.SuccessorList.Nodes[i].IP
		node.Port = int32(s.ChordRing.SuccessorList.Nodes[i].Port)
		response.Nodes = append(response.Nodes, node)
	}
	// TODO store key value in multiple servers
	return response, nil
}

// Get find the value of the given key
func (s *DStoreServer) Get(ctx context.Context, lookup *dstoregrpc.Lookup) (*dstoregrpc.LookupResponse, error) {
	fmt.Printf("SERVER: Getting data %x \n", lookup.Key)
	for _, entity := range s.database {
		fmt.Printf("SERVER: Getting data %+v \n", entity)
	}
	response := &dstoregrpc.LookupResponse{}
	if s.database == nil {
		s.database = make(map[int]*Entity)
	}
	var search [sha256.Size]byte
	copy(search[:sha256.Size], lookup.Key[:sha256.Size])
	for _, entity := range s.database {
		var entityID [sha256.Size]byte
		copy(entityID[:sha256.Size], entity.Identity[:sha256.Size])
		if helpers.Equal(entityID, search) {
			response.Value = []byte(entity.Value)
			return response, nil
		}
	}
	for i := 0; i < len(s.ChordRing.SuccessorList.Nodes); i++ {
		node := &dstoregrpc.Node{}
		node.IP = s.ChordRing.SuccessorList.Nodes[i].IP
		node.Port = int32(s.ChordRing.SuccessorList.Nodes[i].Port)
		response.Nodes = append(response.Nodes, node)
	}
	// TODO store key value in multiple servers
	return response, errors.New("key not found")
}
