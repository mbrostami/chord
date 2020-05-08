package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/mbrostami/chord/internal/app/devsh"
	"github.com/mbrostami/chord/internal/app/devsh/dstore"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "ip address")
	port := flag.Int("port", 0, "port number")
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()
	devshService := devsh.MakeDevsh(*ip, int(*port))
	devshService.StartServer()
	devshService.StartNode()
	if *verbose {
		// go func() {
		// 	for {
		// 		devshService.Chord.Debug()
		// 		time.Sleep(5 * time.Second)
		// 	}
		// }()
	}
	client := dstore.NewClient()
	dstore := dstore.NewStorage(devshService.Chord, client)
	fmt.Println("Enter your username:")
	reader := bufio.NewReader(os.Stdin)
	var cmdString string
	var err error
	for {
		cmdString, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Invalid username, try again %v \n", err)
			continue
		}
		break
	}
	key := cmdString
	value := devshService.Chord.Node.FullAddr()
	numberOfAcceptedNodes, _ := dstore.Store(key, value)
	fmt.Printf("number of nodes written in %d \n", numberOfAcceptedNodes)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter target username:")
		target, _ := reader.ReadString('\n')
		fmt.Printf("Getting value of %s\n", target)
		value, err := dstore.Get(target)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		} else {
			fmt.Printf("Value is %s\n", value)
		}
	}
}
