package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/mbrostami/chord/pkg/chord"
	"github.com/mbrostami/chord/pkg/server"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	flag.Parse()
	var node *chord.Node
	if *port == 0 {
		// Should be a Bootstrap server which is acting like a node
		// with one more functionality to find a closest available node to the newly joining node
		fmt.Print("Bootstrap Node\n")
		node = chord.NewNode("127.0.0.1", 10001)
	} else {
		bootstrapNode := chord.NewNode("127.0.0.1", 10001)
		node = chord.NewNode(*ip, int(*port))
		successor := node.FindRemoteSuccessor(bootstrapNode, node.Identifier)
		node.Join(successor)
	}

	go func() {
		for {
			node.FixFingers() // every 5 seconds
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			node.CheckPredecessor()
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			node.Stabilize()
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			node.Debug()
			time.Sleep(5 * time.Second)
		}
	}()
	server.NewChordServer(node.IP, node.Port, node)
}
