package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/mbrostami/chord"
	"github.com/mbrostami/chord/net"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	flag.Parse()

	remoteSender := net.NewRemoteNodeSenderGrpc()
	var chordRing chord.RingInterface
	var bootstrapNode *chord.RemoteNode

	if *port == 0 {
		// Should be a Bootstrap server which is acting like a node
		// with one more functionality to find a closest available node to the newly joining node
		fmt.Print("Bootstrap Node\n")
		node := chord.NewNode("127.0.0.1", 10001)
		chordRing = chord.NewRing(node, remoteSender)
	} else {
		bootstrapNode = chord.NewRemoteNode(chord.NewNode("127.0.0.1", 10001), remoteSender)
		node := chord.NewNode(*ip, uint(*port))
		chordRing = chord.NewRing(node, remoteSender)
	}

	go net.NewChordReceiver(chordRing)
	time.Sleep(5 * time.Second) // wait until grpc server is up
	if *port != 0 {
		chordRing.Join(bootstrapNode)
	}
	go func() {
		for {
			chordRing.FixFingers() // every 5 seconds
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			chordRing.CheckPredecessor()
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			chordRing.Stabilize()
			time.Sleep(1 * time.Second)
		}
	}()
	if *verbose {
		go func() {
			for {
				chordRing.Verbose()
				time.Sleep(5 * time.Second)
			}
		}()
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
