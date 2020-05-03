package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/mbrostami/chord/pkg/chord"
	"github.com/mbrostami/chord/pkg/server"
)

func main() {
	verbose := flag.Bool("v", false, "verbose")
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	flag.Parse()
	var chordRing *chord.Chord
	if *port == 0 {
		// Should be a Bootstrap server which is acting like a node
		// with one more functionality to find a closest available node to the newly joining node
		fmt.Print("Bootstrap Node\n")
		chordRing = chord.NewChord("127.0.0.1", 10001)
	} else {
		bootstrapNode := &chord.NewChord("127.0.0.1", 10001).Node
		chordRing = chord.NewChord(*ip, int(*port))
		successor := chordRing.FindRemoteSuccessor(bootstrapNode, chordRing.Node.Identifier)
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
