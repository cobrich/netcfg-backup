// Package models defines the data structures used throughout the application.
package models

// Device represents a network device to be backed up.
// It contains connection details, credentials, and the commands to be executed.
type Device struct {
	Host               string   `json:"host"`
	Username           string   `json:"username"`
	Password           string   `json:"password,omitempty"`
	PasswordEnv        string   `json:"password_env,omitempty"`
	KeyPath            string   `json:"key_path,omitempty"`
	Commands           []string `json:"commands"`
	Protocol           string   `json:"protocol"`
	Prompt             string   `json:"prompt,omitempty"`
	TimeoutSeconds     int      `json:"timeout_seconds,omitempty"`
	AllowInsecureAlgos bool     `json:"allow_insecure_algos,omitempty"`
}
