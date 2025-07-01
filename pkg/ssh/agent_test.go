package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSshAgentWithKeyPair(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)
	sshAgent, err := SshAgentWithKeyPair(keyPair)
	assert.NoError(t, err)

	// ensure that socket directory is set in environment, and it exists
	sockFile := filepath.Join(sshAgent.socketDir, "ssh_auth.sock")
	assert.FileExists(t, sockFile)

	// assert that there's 1 key in the agent
	keys, err := sshAgent.agent.List()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)

	sshAgent.Stop()

	// is socketDir removed as expected?
	if _, err := os.Stat(sshAgent.socketDir); !os.IsNotExist(err) {
		assert.FailNow(t, "ssh agent failed to remove socketDir on Stop()")
	}
}

func TestSshAgentWithKeyPairs(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)
	keyPair2, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)
	sshAgent, err := SshAgentWithKeyPairs([]*KeyPair{keyPair, keyPair2})
	assert.NoError(t, err)
	defer sshAgent.Stop()

	keys, err := sshAgent.agent.List()
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestMultipleSshAgents(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)
	keyPair2, err := GenerateRSAKeyPair(2048)
	assert.NoError(t, err)

	// start a couple of agents
	sshAgent, err := SshAgentWithKeyPair(keyPair)
	assert.NoError(t, err)
	sshAgent2, err := SshAgentWithKeyPair(keyPair2)
	assert.NoError(t, err)
	defer sshAgent.Stop()
	defer sshAgent2.Stop()

	// collect public keys from the agents
	keys, err := sshAgent.agent.List()
	assert.NoError(t, err)
	keys2, err := sshAgent2.agent.List()
	assert.NoError(t, err)

	// check that all keys match up to expected
	assert.NotEqual(t, keys, keys2)
	assert.Equal(t, strings.TrimSpace(keyPair.PublicKey), keys[0].String())
	assert.Equal(t, strings.TrimSpace(keyPair2.PublicKey), keys2[0].String())
}
