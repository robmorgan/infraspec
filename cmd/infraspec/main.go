package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/robmorgan/infraspec/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		// Check if this is an ExitError (silent exit with code)
		var exitErr cmd.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
