package connectors

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHConnector struct {
	Host     string
	Username string
	Password string
	KeyPath  string
	Timeout  time.Duration
}

// It creates an authentication method: either by key or by password.
func createAuthMethod(keyPath, password string) (ssh.AuthMethod, error) {
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key from %s: %w", keyPath, err)
		}
		// Create a signer from our key
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// If the key is encrypted with a password, there will be an error here.
			// For simplicity, we only support unencrypted keys for now.
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		return ssh.PublicKeys(signer), nil
	}
	// If the key path is not specified, use the password
	return ssh.Password(password), nil
}

func (s *SSHConnector) RunCommands(cmds []string) ([]models.Result, error) {
	logger := utils.Log.WithField("host", s.Host)
	logger.Infof("SSH: connecting to %s...", s.Host)

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	d := net.Dialer{}
	addr := s.Host
	if !strings.Contains(addr, ":") {
		addr = addr + ":22"
	}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		logger.Errorf("SSH: failed to connect: %v", err)
		return nil, fmt.Errorf("failed to connect to %s: %v", addr, err)
	}
	defer conn.Close()
	logger.Infof("SSH: connection to %s established", addr)

	hostKeyCallback, err := createHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("failed to create HostKeyCallback: %v", err)
	}

	authMethod, err := createAuthMethod(s.KeyPath, s.Password) // <-- s.KeyPath needs to be added to the structure
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            s.Username,
		Auth:            []ssh.AuthMethod{authMethod}, // <-- Use the created method
		HostKeyCallback: hostKeyCallback,
		Timeout:         s.Timeout,
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		logger.Errorf("SSH: failed to create SSH session: %v", err)
		return nil, fmt.Errorf("failed to establish SSH session: %v", err)
	}
	client := ssh.NewClient(c, chans, reqs)
	defer client.Close()

	results := []models.Result{}

	for _, cmd := range cmds {
		logger.Infof("SSH: executing command: %s", cmd)

		outputCh := make(chan []byte)
		errCh := make(chan error)

		go func(c string) {
			session, err := client.NewSession()
			if err != nil {
				errCh <- fmt.Errorf("failed to create session: %v", err)
				return
			}
			defer session.Close()

			output, err := session.CombinedOutput(c)
			if err != nil {
				errCh <- err
				return
			}
			outputCh <- output
		}(cmd)

		select {
		case <-ctx.Done():
			logger.Errorf("SSH: command execution timed out '%s'", cmd)
			results = append(results, models.Result{Cmd: cmd, Output: "timeout"})
		case err := <-errCh:
			logger.Errorf("SSH: error executing command '%s': %v", cmd, err)
			results = append(results, models.Result{Cmd: cmd, Output: fmt.Sprintf("error during execution: %v", err)})
		case output := <-outputCh:
			logger.Infof("SSH: command '%s' executed successfully", cmd)
			results = append(results, models.Result{Cmd: cmd, Output: string(output)})
		}
	}

	return results, nil
}

// New helper function to create HostKeyCallback
func createHostKeyCallback() (ssh.HostKeyCallback, error) {
	// Find the home directory of the current user
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find home directory: %v", err)
	}

	// Form the path to the known_hosts file
	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Create a callback. knownhosts.New will automatically handle file creation if it doesn't exist.
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read known_hosts file (%s): %v", knownHostsPath, err)
	}

	return callback, nil
}
