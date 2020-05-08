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
	response.Nodes = ConvertDstoreToGrpcSuccessorList(s.ChordRing.SuccessorList)
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
	response.Nodes = ConvertDstoreToGrpcSuccessorList(s.ChordRing.SuccessorList)
	if len(response.Nodes) == 0 {
		return nil, errors.New("key not found")
	}
	return response, nil
}
