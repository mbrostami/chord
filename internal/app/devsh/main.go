package devsh

import (
	"fmt"
	"net"
	"time"

	"github.com/mbrostami/chord/internal/clientadapter"
	chordgrpc "github.com/mbrostami/chord/internal/grpc/chord"
	dstoregrpc "github.com/mbrostami/chord/internal/grpc/dstore"
	"github.com/mbrostami/chord/internal/server"
	"github.com/mbrostami/chord/pkg/chord"
	"google.golang.org/grpc"
)

// Devsh struct
type Devsh struct {
	Chord         *chord.Chord
	bootstrap     bool
	bootstrapNode *chord.Node
}

// MakeDevsh create devsh
func MakeDevsh(ip string, port int) *Devsh {
	var client chord.ClientInterface
	client = clientadapter.NewClient()
	devshService := &Devsh{
		bootstrap: false,
	}
	if port == 0 {
		devshService.bootstrap = true
		// Should be a Bootstrap server which is acting like a node
		// with one more functionality to find a closest available node to the newly joining node
		fmt.Print("Bootstrap Node\n")
		devshService.Chord = chord.NewRing("127.0.0.1", 10001, client)
	} else {
		devshService.bootstrapNode = chord.NewNode("127.0.0.1", 10001)
		devshService.Chord = chord.NewRing(ip, port, client)
	}
	return devshService
}

// StartServer start listening as server
func (d *Devsh) StartServer() {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	chordServer := &server.ChordServer{
		ChordRing: d.Chord,
	}
	dstoreServer := &server.DStoreServer{
		ChordRing: d.Chord,
	}
	chordgrpc.RegisterChordServer(grpcServer, chordServer)
	dstoregrpc.RegisterDStoreServer(grpcServer, dstoreServer)
	listener, _ := net.Listen("tcp", d.Chord.Node.FullAddr())
	fmt.Printf("Start listening on makeNodeServer: %s\n", d.Chord.Node.FullAddr())
	go grpcServer.Serve(listener)
	time.Sleep(2 * time.Second) // wait until grpc server is up
}

// StartNode join node to the network and start background services
func (d *Devsh) StartNode() {
	if !d.bootstrap {
		d.Chord.Join(d.bootstrapNode)
	}
	go func() {
		for {
			d.Chord.FixFingers() // every 5 seconds
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			d.Chord.CheckPredecessor()
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			d.Chord.Stabilize()
			time.Sleep(1 * time.Second)
		}
	}()
}
