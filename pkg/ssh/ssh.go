// Package ssh allows to manage SSH connections and send commands through them.
package ssh

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/pkg/retry"
)

// Host is a remote host.
type Host struct {
	Hostname    string // host name or ip address
	SshUserName string // user name
	// set one or more authentication methods,
	// the first valid method will be used
	SshKeyPair       *KeyPair  // ssh key pair to use as authentication method (disabled by default)
	SshAgent         bool      // enable authentication using your existing local SSH agent (disabled by default)
	OverrideSshAgent *SshAgent // enable an in process `SshAgent` for connections to this host (disabled by default)
	Password         string    // plain text password (blank by default)
	CustomPort       int       // port number to use to connect to the host (port 22 will be used if unset)
}

type ScpDownloadOptions struct {
	FileNameFilters []string // File names to match. May include bash-style wildcards. E.g., *.log.
	MaxFileSizeMB   int      // Don't grab any files > MaxFileSizeMB
	RemoteDir       string   // Copy from this directory on the remote machine
	LocalDir        string   // Copy RemoteDir to this directory on the local machine
	RemoteHost      Host     // Connection information for the remote machine
}

// ScpFileTo uploads the contents using SCP to the given host and return an error if the process fails.
func ScpFileTo(host Host, mode os.FileMode, remotePath, contents string) error {
	authMethods, err := createAuthMethodsForHost(host)
	if err != nil {
		return err
	}
	dir, file := filepath.Split(remotePath)

	hostOptions := SshConnectionOptions{
		Username:    host.SshUserName,
		Address:     host.Hostname,
		Port:        host.getPort(),
		Command:     "/usr/bin/scp -t " + dir,
		AuthMethods: authMethods,
	}

	scp := sendScpCommandsToCopyFile(mode, file, contents)

	sshSession := &SshSession{
		Options:  &hostOptions,
		JumpHost: &JumpHostSession{},
		Input:    &scp,
	}

	defer sshSession.Cleanup()

	_, err = runSSHCommand(sshSession)
	return err
}

// ScpFileFrom downloads the file from remotePath on the given host using SCP and returns an error if the process fails.
func ScpFileFrom(host Host, remotePath string, localDestination *os.File, useSudo bool) error {
	authMethods, err := createAuthMethodsForHost(host)
	if err != nil {
		return err
	}

	dir := filepath.Dir(remotePath)

	hostOptions := SshConnectionOptions{
		Username:    host.SshUserName,
		Address:     host.Hostname,
		Port:        host.getPort(),
		Command:     "/usr/bin/scp -t " + dir,
		AuthMethods: authMethods,
	}

	sshSession := &SshSession{
		Options:  &hostOptions,
		JumpHost: &JumpHostSession{},
	}

	defer sshSession.Cleanup()

	return copyFileFromRemote(sshSession, localDestination, remotePath, useSudo)
}

// ScpDirFrom downloads all the files from remotePath on the given host using SCP
// and returns an error if the process fails. NOTE: only files within remotePath will
// be downloaded. This function will not recursively download subdirectories or follow
// symlinks.
func ScpDirFrom(options ScpDownloadOptions, useSudo bool) error {
	authMethods, err := createAuthMethodsForHost(options.RemoteHost)
	if err != nil {
		return err
	}

	hostOptions := SshConnectionOptions{
		Username:    options.RemoteHost.SshUserName,
		Address:     options.RemoteHost.Hostname,
		Port:        options.RemoteHost.getPort(),
		Command:     "/usr/bin/scp -t " + options.RemoteDir,
		AuthMethods: authMethods,
	}

	sshSession := &SshSession{
		Options:  &hostOptions,
		JumpHost: &JumpHostSession{},
	}

	defer sshSession.Cleanup()

	filesInDir, err := listFileInRemoteDir(sshSession, options, useSudo)
	if err != nil {
		return err
	}

	// check if the directory exists
	_, err = os.Stat(options.LocalDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if errors.Is(err, fs.ErrNotExist) {
		err := os.MkdirAll(options.LocalDir, 0o755)
		if err != nil {
			return err
		}
	}

	errorsOccurred := []error{}
	for _, fullRemoteFilePath := range filesInDir {
		fileName := filepath.Base(fullRemoteFilePath)

		localFilePath := filepath.Join(options.LocalDir, fileName)
		localFile, err := os.Create(localFilePath)
		if err != nil {
			return err
		}

		config.Logging.Logger.Infof("Copying remote file: %s to local path %s", fullRemoteFilePath, localFilePath)

		err = copyFileFromRemote(sshSession, localFile, fullRemoteFilePath, useSudo)
		if err != nil {
			errorsOccurred = append(errorsOccurred, err)
		}
	}

	return errors.Join(errorsOccurred...)
}

// CheckSshConnection checks that you can connect via SSH to the given host and return an error if the connection fails.
func CheckSshConnection(host Host) error {
	_, err := CheckSshCommand(host, "'exit'")
	return err
}

// CheckSshConnectionWithRetry attempts to connect via SSH until max retries has been exceeded and returns an error if
// the connection fails
func CheckSshConnectionWithRetry(host Host, retries int, sleepBetweenRetries time.Duration, f ...func(Host) error) error {
	handler := CheckSshConnection
	if f != nil {
		handler = f[0]
	}
	_, err := retry.DoWithRetry(fmt.Sprintf("Checking SSH connection to %s", host.Hostname), retries, sleepBetweenRetries, func() (string, error) {
		return "", handler(host)
	})

	return err
}

// CheckSshCommand checks that you can connect via SSH to the given host and run the given command. Returns the stdout/stderr.
func CheckSshCommand(host Host, command string) (string, error) {
	authMethods, err := createAuthMethodsForHost(host)
	if err != nil {
		return "", err
	}

	hostOptions := SshConnectionOptions{
		Username:    host.SshUserName,
		Address:     host.Hostname,
		Port:        host.getPort(),
		Command:     command,
		AuthMethods: authMethods,
	}

	sshSession := &SshSession{
		Options:  &hostOptions,
		JumpHost: &JumpHostSession{},
	}

	defer sshSession.Cleanup()

	return runSSHCommand(sshSession)
}

// CheckSshCommandWithRetry checks that you can connect via SSH to the given host and run the given command until max retries has been exceeded.
// It return an error if the command fails after max retries has been exceeded.
func CheckSshCommandWithRetry(host Host, command string, retries int, sleepBetweenRetries time.Duration, f ...func(Host, string) (string, error)) (string, error) {
	handler := CheckSshCommand
	if f != nil {
		handler = f[0]
	}
	return retry.DoWithRetry(fmt.Sprintf("Checking SSH connection to %s", host.Hostname), retries, sleepBetweenRetries, func() (string, error) {
		return handler(host, command)
	})
}

// CheckPrivateSshConnection attempts to connect to privateHost (which is not addressable from the Internet) via a
// separate publicHost (which is addressable from the Internet) and then executes "command" on privateHost and returns
// its output. It is useful for checking that it's possible to SSH from a Bastion Host to a private instance.
func CheckPrivateSshConnection(publicHost Host, privateHost Host, command string) (string, error) {
	jumpHostAuthMethods, err := createAuthMethodsForHost(publicHost)
	if err != nil {
		return "", err
	}

	jumpHostOptions := SshConnectionOptions{
		Username:    publicHost.SshUserName,
		Address:     publicHost.Hostname,
		Port:        publicHost.getPort(),
		AuthMethods: jumpHostAuthMethods,
	}

	hostAuthMethods, err := createAuthMethodsForHost(privateHost)
	if err != nil {
		return "", err
	}

	hostOptions := SshConnectionOptions{
		Username:    privateHost.SshUserName,
		Address:     privateHost.Hostname,
		Port:        privateHost.getPort(),
		Command:     command,
		AuthMethods: hostAuthMethods,
		JumpHost:    &jumpHostOptions,
	}

	sshSession := &SshSession{
		Options:  &hostOptions,
		JumpHost: &JumpHostSession{},
	}

	defer sshSession.Cleanup()

	return runSSHCommand(sshSession)
}

// FetchContentsOfFiles connects to the given host via SSH and fetches the contents of the files at the given filePaths.
// If useSudo is true, then the contents will be retrieved using sudo. This method returns a map from file path to
// contents.
func FetchContentsOfFiles(host Host, useSudo bool, filePaths ...string) (map[string]string, error) {
	filePathToContents := map[string]string{}

	for _, filePath := range filePaths {
		contents, err := FetchContentsOfFile(host, useSudo, filePath)
		if err != nil {
			return nil, err
		}

		filePathToContents[filePath] = contents
	}

	return filePathToContents, nil
}

// FetchContentsOfFile connects to the given host via SSH and fetches the contents of the file at the given filePath.
// If useSudo is true, then the contents will be retrieved using sudo. This method returns the contents of that file.
func FetchContentsOfFile(host Host, useSudo bool, filePath string) (string, error) {
	command := fmt.Sprintf("cat %s", filePath)
	if useSudo {
		command = fmt.Sprintf("sudo %s", command)
	}

	return CheckSshCommand(host, command)
}

func listFileInRemoteDir(sshSession *SshSession, options ScpDownloadOptions, useSudo bool) ([]string, error) {
	config.Logging.Logger.Infof("Running command %s on %s@%s", sshSession.Options.Command, sshSession.Options.Username, sshSession.Options.Address)

	var result []string
	var findCommandArgs []string

	if useSudo {
		findCommandArgs = append(findCommandArgs, "sudo")
	}

	findCommandArgs = append(findCommandArgs, "find", options.RemoteDir)
	findCommandArgs = append(findCommandArgs, "-type", "f")

	filtersLength := len(options.FileNameFilters)
	if options.FileNameFilters != nil && filtersLength > 0 {

		findCommandArgs = append(findCommandArgs, "\\(")
		for i, curFilter := range options.FileNameFilters {
			// due to inconsistent bash behavior we need to wrap the
			// filter in single quotes
			curFilter = fmt.Sprintf("'%s'", curFilter)
			findCommandArgs = append(findCommandArgs, "-name", curFilter)

			// only add the or flag if we're not the last element
			if filtersLength-i > 1 {
				findCommandArgs = append(findCommandArgs, "-o")
			}
		}
		findCommandArgs = append(findCommandArgs, "\\)")
	}

	if options.MaxFileSizeMB != 0 {
		findCommandArgs = append(findCommandArgs, "-size", fmt.Sprintf("-%dM", options.MaxFileSizeMB))
	}

	finalCommandString := strings.Join(findCommandArgs, " ")
	resultString, err := CheckSshCommand(options.RemoteHost, finalCommandString)
	if err != nil {
		return result, err
	}

	// The last character returned is `\n` this results in an extra "" array
	// member when we do the split below. Cut off the last character to avoid
	// having to remove the blank entry in the array.
	resultString = resultString[:len(resultString)-1]

	result = append(result, strings.Split(resultString, "\n")...)
	return result, nil
}

// Added based on code: https://github.com/bramvdbogaerde/go-scp/pull/6/files
func copyFileFromRemote(sshSession *SshSession, file *os.File, remotePath string, useSudo bool) error {
	if err := setUpSSHClient(sshSession); err != nil {
		return err
	}

	if err := setUpSSHSession(sshSession); err != nil {
		return err
	}

	command := fmt.Sprintf("dd if=%s", remotePath)
	if useSudo {
		command = fmt.Sprintf("sudo %s", command)
	}

	config.Logging.Logger.Infof("Running command %s on %s@%s", command, sshSession.Options.Username, sshSession.Options.Address)

	r, err := sshSession.Session.Output(command)
	if err != nil {
		fmt.Printf("error reading from remote stdout: %s", err)
	}
	defer sshSession.Session.Close()
	// write to local file
	_, err = file.Write(r)

	return err
}

func runSSHCommand(sshSession *SshSession) (string, error) {
	config.Logging.Logger.Infof("Running command %s on %s@%s", sshSession.Options.Command, sshSession.Options.Username, sshSession.Options.Address)
	if err := setUpSSHClient(sshSession); err != nil {
		return "", err
	}

	if err := setUpSSHSession(sshSession); err != nil {
		return "", err
	}

	if sshSession.Input != nil {
		w, err := sshSession.Session.StdinPipe()
		if err != nil {
			return "", err
		}
		go func() {
			defer w.Close()
			(*sshSession.Input)(w)
		}()
	}

	bytes, err := sshSession.Session.CombinedOutput(sshSession.Options.Command)
	if err != nil {
		return string(bytes), err
	}

	return string(bytes), nil
}

func setUpSSHClient(sshSession *SshSession) error {
	if sshSession.Options.JumpHost == nil {
		return fillSSHClientForHost(sshSession)
	}
	return fillSSHClientForJumpHost(sshSession)
}

func fillSSHClientForHost(sshSession *SshSession) error {
	client, err := createSSHClient(sshSession.Options)
	if err != nil {
		return err
	}

	sshSession.Client = client
	return nil
}

func fillSSHClientForJumpHost(sshSession *SshSession) error {
	jumpHostClient, err := createSSHClient(sshSession.Options.JumpHost)
	if err != nil {
		return err
	}
	sshSession.JumpHost.JumpHostClient = jumpHostClient

	hostVirtualConn, err := jumpHostClient.Dial("tcp", sshSession.Options.ConnectionString())
	if err != nil {
		return err
	}
	sshSession.JumpHost.HostVirtualConnection = hostVirtualConn

	hostConn, hostIncomingChannels, hostIncomingRequests, err := ssh.NewClientConn(hostVirtualConn, sshSession.Options.ConnectionString(), createSSHClientConfig(sshSession.Options))
	if err != nil {
		return err
	}
	sshSession.JumpHost.HostConnection = hostConn

	sshSession.Client = ssh.NewClient(hostConn, hostIncomingChannels, hostIncomingRequests)
	return nil
}

func setUpSSHSession(sshSession *SshSession) error {
	session, err := sshSession.Client.NewSession()
	if err != nil {
		return err
	}

	sshSession.Session = session
	return nil
}

func createSSHClient(options *SshConnectionOptions) (*ssh.Client, error) {
	sshClientConfig := createSSHClientConfig(options)
	return ssh.Dial("tcp", options.ConnectionString(), sshClientConfig)
}

func createSSHClientConfig(hostOptions *SshConnectionOptions) *ssh.ClientConfig {
	clientConfig := &ssh.ClientConfig{
		User: hostOptions.Username,
		Auth: hostOptions.AuthMethods,
		// Do not do a host key check, as Terratest is only used for testing, not prod
		HostKeyCallback: NoOpHostKeyCallback,
		// By default, Go does not impose a timeout, so a SSH connection attempt can hang for a LONG time.
		Timeout: 10 * time.Second,
	}
	clientConfig.SetDefaults()
	return clientConfig
}

// NoOpHostKeyCallback is an ssh.HostKeyCallback that does nothing. Only use this when you're sure you don't want to check the host key at all
// (e.g., only for testing and non-production use cases).
func NoOpHostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}

// Returns an array of authentication methods
func createAuthMethodsForHost(host Host) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// override local ssh agent with given sshAgent instance
	if host.OverrideSshAgent != nil {
		conn, err := net.Dial("unix", host.OverrideSshAgent.socketFile)
		if err != nil {
			fmt.Print("Failed to dial in memory ssh agent")
			return methods, err
		}
		agentClient := agent.NewClient(conn)
		methods = append(methods, []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}...)
	}

	// use existing ssh agent socket
	// if agent authentication is enabled and no agent is set up, returns an error
	if host.SshAgent {
		socket := os.Getenv("SSH_AUTH_SOCK")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return methods, err
		}
		agentClient := agent.NewClient(conn)
		methods = append(methods, []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}...)
	}

	// use provided ssh key pair
	if host.SshKeyPair != nil {
		signer, err := ssh.ParsePrivateKey([]byte(host.SshKeyPair.PrivateKey))
		if err != nil {
			return methods, err
		}
		methods = append(methods, []ssh.AuthMethod{ssh.PublicKeys(signer)}...)
	}

	// Use given password
	if len(host.Password) > 0 {
		methods = append(methods, []ssh.AuthMethod{ssh.Password(host.Password)}...)
	}

	// no valid authentication method was provided
	if len(methods) < 1 {
		return methods, errors.New("no authentication method defined")
	}

	return methods, nil
}

// sendScpCommandsToCopyFile returns a function which will send commands to the SCP binary to output a file on the remote machine.
// A full explanation of the SCP protocol can be found at
// https://web.archive.org/web/20170215184048/https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
func sendScpCommandsToCopyFile(mode os.FileMode, fileName, contents string) func(io.WriteCloser) {
	return func(input io.WriteCloser) {
		octalMode := "0" + strconv.FormatInt(int64(mode), 8)

		// Create a file at <filename> with Unix permissions set to <octalMost> and the file will be <len(content)> bytes long.
		fmt.Fprintln(input, "C"+octalMode, len(contents), fileName)

		// Actually send the file
		fmt.Fprint(input, contents)

		// End of transfer
		fmt.Fprint(input, "\x00")
	}
}

// Gets the port that should be used to communicate with the host
func (h Host) getPort() int {
	// If a CustomPort is not set use standard ssh port
	if h.CustomPort == 0 {
		return 22
	} else {
		return h.CustomPort
	}
}
