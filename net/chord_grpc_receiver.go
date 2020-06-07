package net

import (
	context "context"
	"errors"
	"net"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/mbrostami/chord"
	chordGrpc "github.com/mbrostami/chord/grpc"
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/merkle"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

type ChordGrpcReceiver struct {
	chordGrpc.UnimplementedChordServer
	ring chord.RingInterface
}

func NewChordReceiver(ring chord.RingInterface) *ChordGrpcReceiver {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	chordServer := &ChordGrpcReceiver{
		ring: ring,
	}
	chordGrpc.RegisterChordServer(grpcServer, chordServer)
	listener, _ := net.Listen("tcp", ring.GetLocalNode().GetFullAddress())
	log.Infof("Start listening on makeNodeServer: %s\n", ring.GetLocalNode().GetFullAddress())
	grpcServer.Serve(listener)
	return chordServer
}

// Notify update predecessor
// is being called periodically
func (s *ChordGrpcReceiver) Notify(ctx context.Context, caller *chordGrpc.Node) (*wrappers.BoolValue, error) {
	result := &wrappers.BoolValue{
		Value: s.ring.Notify(chordGrpc.ConvertToChordNode(caller)),
	}
	return result, nil
}

// GetStablizerData get predecessor node + successor list
func (s *ChordGrpcReceiver) GetStablizerData(ctx context.Context, caller *chordGrpc.Node) (*chordGrpc.StablizerData, error) {
	stabilizerData := &chordGrpc.StablizerData{}
	predecessor, successorList := s.ring.GetStabilizerData(chordGrpc.ConvertToChordNode(caller))
	stabilizerData.Predecessor = chordGrpc.ConvertToGrpcNode(predecessor.Node)
	stabilizerData.SuccessorList = chordGrpc.ConvertToGrpcSuccessorList(successorList)
	return stabilizerData, nil
}

// FindSuccessor get closest node to the given key
func (s *ChordGrpcReceiver) FindSuccessor(ctx context.Context, lookup *chordGrpc.Lookup) (*chordGrpc.Node, error) {
	successor := s.ring.FindSuccessor(helpers.ConvertToHashSized(lookup.Key))
	if successor == nil {
		log.Error("receiver.FindSuccessor: Successor is null")
		return nil, errors.New("successor is null")
	}
	return chordGrpc.ConvertToGrpcNode(successor.Node), nil
}

// Store store data in database
func (s *ChordGrpcReceiver) Store(ctx context.Context, content *chordGrpc.Content) (*wrappers.BoolValue, error) {
	stored := s.ring.Store(content.Data)
	result := &wrappers.BoolValue{
		Value: stored,
	}
	return result, nil
}

// GetPredecessorList get predecessor list
func (s *ChordGrpcReceiver) GetPredecessorList(ctx context.Context, caller *chordGrpc.Node) (*chordGrpc.Nodes, error) {
	pList := s.ring.GetPredecessorList(chordGrpc.ConvertToChordNode(caller))
	nodes := &chordGrpc.Nodes{
		Nodes: chordGrpc.ConvertToGrpcPredecessorList(pList),
	}
	return nodes, nil
}

// GlobalMaintenance to sync data from predecessor
func (s *ChordGrpcReceiver) GlobalMaintenance(ctx context.Context, replicationRequest *chordGrpc.Replication) (*chordGrpc.Replication, error) {
	replicationResponse, err := s.ring.GlobalMaintenance(replicationRequest.Data)
	return &chordGrpc.Replication{Data: replicationResponse}, err
}

// ForwardSync sync data from predecessor call
func (s *ChordGrpcReceiver) ForwardSync(ctx context.Context, syncData *chordGrpc.ForwardSyncData) (*chordGrpc.ForwardSyncData, error) {

	var serializedNodes []*merkle.SerializedNode
	for i := 0; i < len(syncData.MerkleTree.Nodes); i++ {
		serializedNode := &merkle.SerializedNode{
			Hash: syncData.MerkleTree.Nodes[i].Hash,
		}
		serializedNodes = append(serializedNodes, serializedNode)
	}
	var identifier [helpers.HashSize]byte
	copy(identifier[:helpers.HashSize], syncData.PredecessorListHash[:helpers.HashSize])

	diffNodes, err := s.ring.ForwardSync(syncData.Data, identifier, serializedNodes)
	if err != nil {
		log.Error("receiver.ForwardSync: forward sync failed! %v", err)
		return nil, err
	}
	if diffNodes == nil {
		// replicated successfully
		return nil, nil
	}
	// there are some diffs in merkle tree which are not synced

	grpcMerkleTree := &chordGrpc.MerkleTree{}
	for _, node := range diffNodes {
		grpcMerkleTree.Nodes = append(grpcMerkleTree.Nodes, &chordGrpc.MerkleNode{
			Hash: node.Hash,
		})
	}
	responseSyncData := &chordGrpc.ForwardSyncData{
		MerkleTree: grpcMerkleTree,
	}
	return responseSyncData, nil
}
