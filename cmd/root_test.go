package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			wantErr:     true,
			errContains: "requires at least 1 arg(s)",
		},
		// Note: "too many arguments" test removed - we now accept multiple feature paths/directories
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			RootCmd.SetOut(buf)
			RootCmd.SetArgs(tt.args)

			err := RootCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	assert.False(t, verbose)
	RootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "verbose" {
			assert.Equal(t, "false", flag.DefValue)
		}
	})
}

func TestParallelFlags(t *testing.T) {
	// Check parallel flag exists with correct default
	parallelFlag := RootCmd.PersistentFlags().Lookup("parallel")
	assert.NotNil(t, parallelFlag)
	assert.Equal(t, "0", parallelFlag.DefValue)
	assert.Equal(t, "p", parallelFlag.Shorthand)

	// Check timeout flag exists with correct default
	timeoutFlag := RootCmd.PersistentFlags().Lookup("timeout")
	assert.NotNil(t, timeoutFlag)
	assert.Equal(t, "0", timeoutFlag.DefValue)
}
