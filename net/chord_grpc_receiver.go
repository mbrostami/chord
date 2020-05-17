package net

import (
	context "context"
	"errors"
	fmt "fmt"
	"net"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/mbrostami/chord"
	chordGrpc "github.com/mbrostami/chord/grpc"
	"github.com/mbrostami/chord/helpers"
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
	fmt.Printf("Start listening on makeNodeServer: %s\n", ring.GetLocalNode().GetFullAddress())
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
	stabilizerData.Predecessor = chordGrpc.ConvertToGrpcNode(predecessor)
	stabilizerData.SuccessorList = chordGrpc.ConvertToGrpcSuccessorList(successorList)
	return stabilizerData, nil
}

// FindSuccessor get closest node to the given key
func (s *ChordGrpcReceiver) FindSuccessor(ctx context.Context, lookup *chordGrpc.Lookup) (*chordGrpc.Node, error) {
	successor := s.ring.FindSuccessor(helpers.ConvertToHashSized(lookup.Key))
	if successor == nil {
		return nil, errors.New("successor is null")
	}
	return chordGrpc.ConvertToGrpcNode(successor.Node), nil
}
