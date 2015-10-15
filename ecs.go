package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/gawkermedia/ecs/cluster"
	"github.com/gawkermedia/ecs/task"
	"os"
)

var az = "us-east-1"

func printHelp() {
	fmt.Fprintf(os.Stdout, "Usage: "+os.Args[0]+" command [parameters]\n")
	fmt.Fprintf(os.Stdout, "Help: "+os.Args[0]+" help [command]\n")
	fmt.Fprintf(os.Stdout, "Available commands: cluster task\n")
}

// ecs cmd [args...]
// ecs help cmd
func main() {
	defaults.DefaultConfig.Region = aws.String(az)

	var cmd string = "help"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
		if cmd == "help" && len(os.Args) == 3 {
			cmd = os.Args[2]
		}
	}

	var ret []*string
	var err error = nil

	switch {
	case cmd == "cluster":
		ret, err = cluster.Run(cmd, os.Args[2:])
	case cmd == "task":
		ret, err = task.Run(cmd, os.Args[2:])
	case cmd == "help":
		printHelp()
		return
	default:
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stdout, err.Error()+"\n")
		os.Exit(1)
	}
	for _, v := range ret {
		fmt.Println(*v)
	}

}
