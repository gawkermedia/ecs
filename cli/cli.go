package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
)

type (
	// Func The operation
	Func func(*ecs.ECS, []string) ([]*string, error)
	// HelpFunc The desciption of the CLI operation
	HelpFunc func([]string) *flag.FlagSet
)

// Command A subcommand definition
type Command struct {
	Cmd  Func
	Desc string
	Help HelpFunc
}

// Get Returns a new command line parser
func Get(name string, args []string) *flag.FlagSet {
	var cli = flag.NewFlagSet(name, flag.ExitOnError)
	return cli
}

// String if val is empty returns nil else returns a pointer to val
func String(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}

// Failure Returns the failures as an error.It returns a nil if there are no failures.
func Failure(failures []*ecs.Failure, err error) error {
	if len(failures) == 0 {
		return err
	}

	var failMessages = make([]string, len(failures))
	for i, v := range failures {
		failMessages[i] = "failure reason: " + *v.Reason + " (" + *v.Arn + ")"
	}
	return errors.New(strings.Join(failMessages, "\n"))
}

// PrintHelp Display an usage message.
func PrintHelp(cmd string, commands map[string]Command, args []string) {
	fmt.Fprintf(os.Stderr, "Available "+cmd+" subcommands:\n")
	for k := range commands {
		if cmd, ok := commands[k]; ok {
			fmt.Fprintf(os.Stderr, "\n  # "+k+"\n  "+cmd.Desc+"\n\n  Parameters:\n")
			ret := cmd.Help(args)
			ret.PrintDefaults()
		}
	}
}

// Run Main entry point, which runs a command or display a help message.
func Run(command string, commands map[string]Command, args []string) ([]*string, error) {
	svc := ecs.New(nil)

	var input string
	if len(args) > 0 {
		input = args[0]
	}

	if cmd, ok := commands[input]; ok {
		if len(args) > 1 && args[1] == "help" {
			fmt.Fprintf(os.Stderr, "\n  # "+input+"\n  "+cmd.Desc+"\n\n  Parameters:\n")
			cmd.Help(args).PrintDefaults()
			return nil, nil
		}
		ret, err := cmd.Cmd(svc, args[1:]) //func(*ecs.ECS, []string) ([]*string, error))(svc, args[1:])
		return ret, err
	}
	PrintHelp(command, commands, args)
	return nil, nil
}
