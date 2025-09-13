package connectors

import (
	"fmt"
	"net" // CHANGE: needed for DialTimeout
	"strings"
	"time"

	"github.com/cobrich/netcfg-backup/models"
	"github.com/cobrich/netcfg-backup/utils"

	"github.com/ziutek/telnet"
)

// TelnetConnector implements the Connector interface for the Telnet protocol.

type TelnetConnector struct {
	Host     string
	Username string
	Password string
	Prompt   string
	Timeout  time.Duration // This timeout will now be for every operation
}

// Set reasonable default timeouts
const defaultTelnetTimeout = 15 * time.Second

// RunCommands connects to a device via Telnet and executes a list of commands.
func (t *TelnetConnector) RunCommands(cmds []string) ([]models.Result, error) {
	logger := utils.Log.WithField("host", t.Host)
	logger.Infof("Telnet: connecting to %s...", t.Host)

	// CHANGE: Use net.DialTimeout to connect with a timeout
	addr := t.Host
	if !strings.Contains(addr, ":") {
		addr = addr + ":23"
	}
	connDialer, err := net.DialTimeout("tcp", addr, t.getTimeout())
	if err != nil {
		logger.Errorf("Telnet: failed to connect: %v", err)
		return nil, fmt.Errorf("telnet: failed to connect: %v", err)
	}

	// Wrap the connection in telnet.Conn
	conn, err := telnet.NewConn(connDialer)
	if err != nil {
		// This error can occur if the Telnet handshake (option exchange) fails
		logger.Errorf("Telnet: failed to create Telnet session: %v", err)
		connDialer.Close() // Close the base TCP connection
		return nil, fmt.Errorf("telnet: failed to create Telnet session: %v", err)
	}

	defer conn.Close()

	// Set a deadline for the entire connection. It will be shifted for each operation
	conn.SetUnixWriteMode(true)

	// --- Authorization with deadlines ---
	if err := expect(conn, t.getTimeout(), "Username:", "login:"); err != nil {
		return nil, fmt.Errorf("telnet: did not wait for username prompt: %v", err)
	}
	if err := send(conn, t.getTimeout(), t.Username); err != nil {
		return nil, fmt.Errorf("telnet: failed to send username: %v", err)
	}

	if err := expect(conn, t.getTimeout(), "Password:", "password:"); err != nil {
		return nil, fmt.Errorf("telnet: did not wait for password prompt: %v", err)
	}
	if err := send(conn, t.getTimeout(), t.Password); err != nil {
		return nil, fmt.Errorf("telnet: failed to send password: %v", err)
	}

	if t.Prompt == "" {
		t.Prompt = ">"
	}
	// Wait for the prompt after login
	if _, err := readUntil(conn, t.getTimeout(), t.Prompt); err != nil {
		return nil, fmt.Errorf("telnet: did not wait for prompt after login: %v", err)
	}

	results := []models.Result{}

	for _, cmd := range cmds {
		logger.Infof("Telnet: executing command: %s", cmd)

		if err := send(conn, t.getTimeout(), cmd); err != nil {
			results = append(results, models.Result{Cmd: cmd, Output: fmt.Sprintf("error sending: %v", err)})
			continue // Go to the next command
		}

		output, err := readUntil(conn, t.getTimeout(), t.Prompt)
		if err != nil {
			logger.Errorf("Telnet: error executing command '%s': %v", cmd, err)
			results = append(results, models.Result{Cmd: cmd, Output: fmt.Sprintf("error during execution: %v", err)})
		} else {
			// Telnet output often includes the command that was just typed and the prompt that follows the output.
			// This function cleans up the raw output to return only the actual command response.
			cleanOutput := cleanTelnetOutput(output, cmd, t.Prompt)
			logger.Infof("Telnet: command '%s' executed successfully", cmd)
			results = append(results, models.Result{Cmd: cmd, Output: cleanOutput})
		}
	}

	return results, nil
}

// getTimeout returns the configured timeout or a default value.
func (t *TelnetConnector) getTimeout() time.Duration {
	if t.Timeout > 0 {
		return t.Timeout
	}
	return defaultTelnetTimeout
}

// expect reads from the connection until one of the delimiters is found.
func expect(conn *telnet.Conn, timeout time.Duration, delimiters ...string) error {
	conn.SetReadDeadline(time.Now().Add(timeout))
	return conn.SkipUntil(delimiters...)
}

// send writes a string to the connection with a timeout.
func send(conn *telnet.Conn, timeout time.Duration, s string) error {
	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write([]byte(s + "\n"))
	return err
}

// readUntil reads from the connection until the prompt is found.
func readUntil(conn *telnet.Conn, timeout time.Duration, prompt string) (string, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	data, err := conn.ReadUntil(prompt)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// cleanTelnetOutput removes the command echo and prompt from the output.
func cleanTelnetOutput(output, cmd, prompt string) string {
	// Remove the command echo (if it is at the beginning)
	output = strings.TrimPrefix(output, cmd)
	// Remove the prompt (if it is at the end)
	output = strings.TrimSuffix(output, prompt)
	// Remove extra spaces and line breaks at the edges
	return strings.TrimSpace(output)
}
