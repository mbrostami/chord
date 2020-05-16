package main

import (
	"flag"
	"github.com/mbrostami/chord"
)

func main() {
	// verbose := flag.Bool("v", false, "verbose")
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	flag.Parse()

	remoteSender := chord.NewRemoteSender()
	bootstrapNode = chord.NewRemoteNode("127.0.0.1", 10001, remoteSender)
	chordRing = chord.NewNode(*ip, uint(*port))

}
