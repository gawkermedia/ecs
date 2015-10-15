package cli

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ecs"
	"os"
	"strings"
)

type (
	CliFunc  func(*ecs.ECS, []string) ([]*string, error)
	HelpFunc func([]string) *flag.FlagSet
)

type Command struct {
	Cmd  CliFunc
	Desc string
	Help HelpFunc
}

func Get(name string, args []string) *flag.FlagSet {
	var cli = flag.NewFlagSet(name, flag.ExitOnError)
	return cli
}

func String(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}

func Failure(failures []*ecs.Failure, err error) error {
	if len(failures) == 0 {
		return err
	}

	var failMessages []string = make([]string, len(failures))
	for i, v := range failures {
		failMessages[i] = "failure reason: " + *v.Reason + " (" + *v.Arn + ")"
	}
	return errors.New(strings.Join(failMessages, "\n"))
}

func PrintHelp(cmd string, commands map[string]Command, args []string) ([]*string, error) {
	fmt.Fprintf(os.Stderr, "Available "+cmd+" subcommands:\n")
	for k := range commands {
		if cmd, ok := commands[k]; ok {
			fmt.Fprintf(os.Stderr, "\n  # "+k+"\n  "+cmd.Desc+"\n\n  Parameters:\n")
			ret := cmd.Help(args)
			ret.PrintDefaults()
		}
	}
	return nil, nil
}

func Run(command string, commands map[string]Command, args []string) ([]*string, error) {
	svc := ecs.New(nil)

	var input string
	if len(args) > 0 {
		input = args[0]
	}

	if cmd, ok := commands[input]; ok {
		ret, err := cmd.Cmd(svc, args[1:]) //func(*ecs.ECS, []string) ([]*string, error))(svc, args[1:])
		return ret, err
	} else {
		PrintHelp(command, commands, args)
		return nil, nil
	}
}
