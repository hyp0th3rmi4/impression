package main

import (
	"github.com/hyp0th3rmi4/impression/cmd"
	"github.com/spf13/cobra"
)

// main creates and executes a command that creates
// the impression engine and runs it with the parameters
// configured via the command line.
func main() {

	cmd := cmd.NewExecuteCommand()
	cobra.CheckErr(cmd.Execute())
}
