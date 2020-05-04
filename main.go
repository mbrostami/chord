package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/mbrostami/chord/internal/clientadapter"
	"github.com/mbrostami/chord/internal/server"
	"github.com/mbrostami/chord/pkg/chord"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	flag.Parse()
	var chordRing *chord.Chord

	var client chord.ClientInterface
	client = clientadapter.NewClient()

	if *port == 0 {
		// Should be a Bootstrap server which is acting like a node
		// with one more functionality to find a closest available node to the newly joining node
		fmt.Print("Bootstrap Node\n")
		chordRing = chord.NewRing("127.0.0.1", 10001, client)
	} else {
		bootstrapNode := chord.NewNode("127.0.0.1", 10001)
		chordRing = chord.NewRing(*ip, int(*port), client)
		successor, err := client.FindSuccessor(bootstrapNode, chordRing.Node.Identifier[:])
		if err != nil {
			panic("bootstrap node is not available")
		}
		chordRing.Join(successor)
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
				chordRing.Debug()
				time.Sleep(5 * time.Second)
			}
		}()
	}
	server.NewChordServer(chordRing)
}
