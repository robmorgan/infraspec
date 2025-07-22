package ssh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostWithDefaultPort(t *testing.T) {
	t.Parallel()

	host := &Host{}

	assert.Equal(t, 22, host.getPort(), "host.getPort() did not return the default ssh port of 22")
}

func TestHostWithCustomPort(t *testing.T) {
	t.Parallel()

	customPort := 2222
	host := &Host{CustomPort: customPort}

	assert.Equal(t, customPort, host.getPort(), "host.getPort() did not return the custom port number")
}

// global var for use in mock callback
var timesCalled int

func TestCheckSshConnectionWithRetry(t *testing.T) {
	// Reset the global call count
	timesCalled = 0

	host := &Host{Hostname: "Host"}
	retries := 10

	assert.Nil(t, CheckSshConnectionWithRetry(host, retries, 3, mockSshConnection))
}

func TestCheckSshConnectionWithRetryExceedsMaxRetries(t *testing.T) {
	// Reset the global call count
	timesCalled = 0

	host := &Host{Hostname: "Host"}

	// Not enough retries
	retries := 3

	assert.Error(t, CheckSshConnectionWithRetry(host, retries, 3, mockSshConnection))
}

func TestCheckSshCommandWithRetry(t *testing.T) {
	// Reset the global call count
	timesCalled = 0

	host := &Host{Hostname: "Host"}
	command := "echo -n hello world"
	retries := 10

	_, err := CheckSshCommandWithRetry(host, command, retries, 3, mockSshCommand)
	assert.Nil(t, err)
}

func TestCheckSshCommandWithRetryExceedsRetries(t *testing.T) {
	// Reset the global call count
	timesCalled = 0

	host := &Host{Hostname: "Host"}
	command := "echo -n hello world"

	// Not enough retries
	retries := 3

	_, err := CheckSshCommandWithRetry(host, command, retries, 3, mockSshCommand)
	assert.Error(t, err)
}

func mockSshCommand(host *Host, command string) (string, error) {
	return "", mockSshConnection(host)
}

func mockSshConnection(host *Host) error {
	timesCalled += 1
	if timesCalled >= 5 {
		return nil
	} else {
		return fmt.Errorf("Called %v times", timesCalled)
	}
}
