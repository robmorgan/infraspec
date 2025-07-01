package ssh

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/agent"

	"github.com/robmorgan/infraspec/internal/config"
)

type SshAgent struct {
	stop       chan bool
	stopped    chan bool
	socketDir  string
	socketFile string
	agent      agent.Agent
	ln         net.Listener
}

// Create SSH agent, start it in background and returns control back to the main thread
// You should stop the agent to cleanup files afterwards by calling `defer s.Stop()`
func NewSshAgent(socketDir string, socketFile string) (*SshAgent, error) {
	var err error
	s := &SshAgent{make(chan bool), make(chan bool), socketDir, socketFile, agent.NewKeyring(), nil}
	s.ln, err = net.Listen("unix", s.socketFile)
	if err != nil {
		return nil, err
	}
	go s.run()
	return s, nil
}

// expose socketFile variable
func (s *SshAgent) SocketFile() string {
	return s.socketFile
}

// SSH Agent listener and handler
func (s *SshAgent) run() {
	defer close(s.stopped)
	for {
		select {
		case <-s.stop:
			return
		default:
			c, err := s.ln.Accept()
			if err != nil {
				select {
				// When s.Stop() closes the listener, s.ln.Accept() returns an error that can be ignored
				// since the agent is in stopping process
				case <-s.stop:
					return
					// When s.ln.Accept() returns a legit error, we print it and continue accepting further requests
				default:
					config.Logging.Logger.Errorf("could not accept connection to agent %v", err)
					continue
				}
			} else {
				defer c.Close()
				go func(c io.ReadWriter) {
					err := agent.ServeAgent(s.agent, c)
					if err != nil {
						config.Logging.Logger.Errorf("could not serve ssh agent %v", err)
					}
				}(c)
			}
		}
	}
}

// Stop and clean up SSH agent
func (s *SshAgent) Stop() {
	close(s.stop)
	s.ln.Close()
	<-s.stopped
	os.RemoveAll(s.socketDir)
}

// Instantiates and returns an in-memory ssh agent with the given KeyPair already added
// You should stop the agent to cleanup files afterwards by calling `defer sshAgent.Stop()`
func SshAgentWithKeyPair(keyPair *KeyPair) (*SshAgent, error) {
	sshAgent, err := SshAgentWithKeyPairs([]*KeyPair{keyPair})
	if err != nil {
		return nil, err
	}

	return sshAgent, nil
}

// SshAgentWithKeyPairs Instantiates and returns an in-memory ssh agent with the given KeyPair(s) already added
// You should stop the agent to cleanup files afterwards by calling `defer sshAgent.Stop()`
func SshAgentWithKeyPairs(keyPairs []*KeyPair) (*SshAgent, error) {
	// Instantiate a temporary SSH agent
	socketDir, err := os.MkdirTemp("", "ssh-agent-")
	if err != nil {
		return nil, err
	}
	socketFile := filepath.Join(socketDir, "ssh_auth.sock")
	sshAgent, err := NewSshAgent(socketDir, socketFile)
	if err != nil {
		return nil, err
	}

	// add given ssh keys to the newly created agent
	for _, keyPair := range keyPairs {
		// Create SSH key for the agent using the given SSH key pair(s)
		block, _ := pem.Decode([]byte(keyPair.PrivateKey))
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key := agent.AddedKey{PrivateKey: privateKey}
		sshAgent.agent.Add(key)
	}

	return sshAgent, err
}
