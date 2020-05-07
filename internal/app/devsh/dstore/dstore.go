package dstore

import (
	"github.com/mbrostami/chord/pkg/chord"
	"github.com/mbrostami/chord/pkg/helpers"
)

// DStore distributed storage
type DStore struct {
	chord  *chord.Chord
	client *GrpcClient
}

// NewStorage creates new distributed storage
func NewStorage(chord *chord.Chord, client *GrpcClient) *DStore {
	return &DStore{chord: chord, client: client}
}

// Store data in remote node
func (ds *DStore) Store(key string, value string) (int, error) {
	hashKey := helpers.Hash(key)
	nodeToStore := ds.chord.FindSuccessor(hashKey)
	return ds.client.Store(nodeToStore, hashKey, []byte(value))
}

// Get data
func (ds *DStore) Get(key string) (string, error) {
	hashKey := helpers.Hash(key)
	nodeToFetch := ds.chord.FindSuccessor(hashKey)
	value, err := ds.client.Get(nodeToFetch, hashKey)
	if err != nil {
		return "", err
	}
	return string(value), nil
}
