package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunInit(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func()
		cleanupFunc func()
		wantErr     bool
		wantOutput  []string
	}{
		{
			name: "creates features directory when it doesn't exist",
			setupFunc: func() {
				os.RemoveAll("./features")
			},
			cleanupFunc: func() {
				os.RemoveAll("./features")
			},
			wantErr: false,
			wantOutput: []string{
				"🎉 Successfully initialized InfraSpec!",
				"📁 Created features directory at:",
				"Next steps:",
				"1. Create your first test: infraspec new my-test.feature",
				"2. Run your tests: infraspec features/my-test.feature",
			},
		},
		{
			name: "handles existing features directory",
			setupFunc: func() {
				os.MkdirAll("./features", 0755)
			},
			cleanupFunc: func() {
				os.RemoveAll("./features")
			},
			wantErr: false,
			wantOutput: []string{
				"✅ Features directory already exists at ./features",
			},
		},
		{
			name: "handles directory creation error",
			setupFunc: func() {
				os.RemoveAll("./features")
				// Create a file named 'features' to cause directory creation to fail
				os.WriteFile("./features", []byte("test"), 0644)
			},
			cleanupFunc: func() {
				os.RemoveAll("./features")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}
			defer func() {
				if tt.cleanupFunc != nil {
					tt.cleanupFunc()
				}
			}()

			buf := new(bytes.Buffer)
			cmd := &cobra.Command{}
			cmd.SetOut(buf)

			err := runInit(cmd, []string{})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				output := buf.String()
				for _, expectedOutput := range tt.wantOutput {
					assert.Contains(t, output, expectedOutput)
				}
			}

			if !tt.wantErr {
				// Verify directory exists and has correct permissions
				info, err := os.Stat("./features")
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
				assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
			}
		})
	}
}

func TestInitCommand(t *testing.T) {
	assert.Equal(t, "init", initCmd.Use)
	assert.Equal(t, "Initialize InfraSpec in the current directory", initCmd.Short)
	assert.Contains(t, initCmd.Long, "Initialize InfraSpec in the current directory")
	assert.NotNil(t, initCmd.RunE)
}
