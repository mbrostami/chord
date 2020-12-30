package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mbrostami/chord"
	"github.com/mbrostami/chord/helpers"
	"github.com/mbrostami/chord/net"
	log "github.com/sirupsen/logrus"
)

func main() {
	logLevelWarning := flag.Bool("v", false, "verbose (warning)")
	logLevelInfo := flag.Bool("vv", false, "verbose (info)")
	logLevelDebug := flag.Bool("vvv", false, "verbose (debug)")
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	flag.Parse()

	if *logLevelDebug {
		log.SetLevel(log.DebugLevel)
	} else if *logLevelInfo {
		log.SetLevel(log.InfoLevel)
	} else if *logLevelWarning {
		log.SetLevel(log.WarnLevel)
	}

	remoteSender := net.NewRemoteNodeSenderGrpc()
	var chordRing chord.RingInterface
	var bootstrapNode *chord.RemoteNode

	if *port == 0 {
		// Should be a Bootstrap server which is acting like a node
		// with one more functionality to find a closest available node to the newly joining node
		log.Info("Bootstrap Node")
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
			chordRing.FixFingers()
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
	go func() {
		for {
			chordRing.SyncData()
			time.Sleep(10 * time.Second)
		}
	}()
	log.Debugf("Current Node: %x", chordRing.GetLocalNode().Identifier)
	go func() {
		for {
			chordRing.Verbose()
			time.Sleep(5 * time.Second)
		}
	}()
	if *port != 0 {
		reader := bufio.NewReader(os.Stdin)
		for {
			log.Debugf("Current Node: %x", chordRing.GetLocalNode().Identifier)
			fmt.Print("Enter value to store: ")
			username, _ := reader.ReadString('\n')
			record := &chord.Record{
				CreationTime: time.Now(),
				Content:      []byte(username),
				Identifier:   helpers.Hash(username),
			}
			remoteNodeToStore := chordRing.FindSuccessor(record.Hash())
			remoteNodeToStore.Store(record.GetJson())
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
