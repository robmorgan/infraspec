package main

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
			errContains: "accepts 1 arg(s), received 0",
		},
		{
			name:        "too many arguments",
			args:        []string{"test1", "test2"},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

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
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "verbose" {
			assert.Equal(t, "false", flag.DefValue)
		}
	})
}
