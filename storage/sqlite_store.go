package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cobrich/netcfg-backup/models"
	_ "github.com/mattn/go-sqlite3" // The blank import for the driver
)

// ErrDeviceExists is a custom error type for duplicate devices.
type ErrDeviceExists struct {
	Host string
}

func (e *ErrDeviceExists) Error() string {
	return fmt.Sprintf("device with host '%s' already exists", e.Host)
}

// SQLiteStore implements the Store interface using a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLiteStore and ensures the database schema is set up.
func NewSQLiteStore(filePath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Create the devices table if it doesn't exist.
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return store, nil
}

// initSchema creates the necessary tables in the database.
func (s *SQLiteStore) initSchema() error {
	query := `
    CREATE TABLE IF NOT EXISTS devices (
        host TEXT NOT NULL PRIMARY KEY,
        username TEXT NOT NULL,
        password TEXT,
        password_env TEXT,
        key_path TEXT,
        commands TEXT, -- Storing commands as a JSON array string
        protocol TEXT NOT NULL,
        prompt TEXT,
        timeout_seconds INTEGER,
        allow_insecure_algos BOOLEAN
    );`

	_, err := s.db.Exec(query)
	return err
}

// GetAllDevices retrieves all devices from the database.
func (s *SQLiteStore) GetAllDevices() ([]models.Device, error) {
	rows, err := s.db.Query("SELECT host, username, password, password_env, key_path, commands, protocol, prompt, timeout_seconds, allow_insecure_algos FROM devices ORDER BY host")
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var dev models.Device
		var commandsJSON string // We'll read the JSON string here

		err := rows.Scan(
			&dev.Host, &dev.Username, &dev.Password, &dev.PasswordEnv,
			&dev.KeyPath, &commandsJSON, &dev.Protocol, &dev.Prompt,
			&dev.TimeoutSeconds, &dev.AllowInsecureAlgos,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}

		// Convert the JSON string of commands back to a string slice
		if err := json.Unmarshal([]byte(commandsJSON), &dev.Commands); err != nil {
			return nil, fmt.Errorf("failed to unmarshal commands for host %s: %w", dev.Host, err)
		}

		devices = append(devices, dev)
	}

	return devices, nil
}

// GetDeviceByHost finds a single device by its host.
func (s *SQLiteStore) GetDeviceByHost(host string) (*models.Device, error) {
	row := s.db.QueryRow("SELECT host, username, password, password_env, key_path, commands, protocol, prompt, timeout_seconds, allow_insecure_algos FROM devices WHERE host = ?", host)

	var dev models.Device
	var commandsJSON string

	err := row.Scan(
		&dev.Host, &dev.Username, &dev.Password, &dev.PasswordEnv,
		&dev.KeyPath, &commandsJSON, &dev.Protocol, &dev.Prompt,
		&dev.TimeoutSeconds, &dev.AllowInsecureAlgos,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device with host '%s' not found", host)
		}
		return nil, fmt.Errorf("failed to scan device row: %w", err)
	}

	if err := json.Unmarshal([]byte(commandsJSON), &dev.Commands); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commands for host %s: %w", dev.Host, err)
	}

	return &dev, nil
}

// AddDevice adds a new device to the database.
func (s *SQLiteStore) AddDevice(dev models.Device) error {
	// Convert commands slice to a JSON string for storage.
	commandsJSON, err := json.Marshal(dev.Commands)
	if err != nil {
		return fmt.Errorf("failed to marshal commands to JSON: %w", err)
	}

	query := `
    INSERT INTO devices (host, username, password, password_env, key_path, commands, protocol, prompt, timeout_seconds, allow_insecure_algos)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err = s.db.Exec(query,
		dev.Host, dev.Username, dev.Password, dev.PasswordEnv,
		dev.KeyPath, string(commandsJSON), dev.Protocol, dev.Prompt,
		dev.TimeoutSeconds, dev.AllowInsecureAlgos,
	)

	// Check for unique constraint violation (duplicate host)
	if err != nil && err.Error() == "UNIQUE constraint failed: devices.host" {
		return &ErrDeviceExists{Host: dev.Host}
	}

	return err
}

// UpdateDevice updates an existing device in the database.
func (s *SQLiteStore) UpdateDevice(dev models.Device) error {
	commandsJSON, err := json.Marshal(dev.Commands)
	if err != nil {
		return fmt.Errorf("failed to marshal commands to JSON: %w", err)
	}

	query := `
    UPDATE devices SET
        username = ?, password = ?, password_env = ?, key_path = ?, commands = ?,
        protocol = ?, prompt = ?, timeout_seconds = ?, allow_insecure_algos = ?
    WHERE host = ?;`

	res, err := s.db.Exec(query,
		dev.Username, dev.Password, dev.PasswordEnv, dev.KeyPath, string(commandsJSON),
		dev.Protocol, dev.Prompt, dev.TimeoutSeconds, dev.AllowInsecureAlgos,
		dev.Host, // This is for the WHERE clause
	)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("device with host '%s' not found to update", dev.Host)
	}

	return err
}

// RemoveDevice removes a device from the database by its host.
func (s *SQLiteStore) RemoveDevice(host string) error {
	res, err := s.db.Exec("DELETE FROM devices WHERE host = ?", host)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("device with host '%s' not found", host)
	}

	return err
}
