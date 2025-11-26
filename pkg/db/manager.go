package db

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/FreePeak/infra-mcp-server/pkg/logger"
)

// DatabaseConnectionConfig represents a single database connection configuration
type DatabaseConnectionConfig struct {
	ID       string `json:"id"`   // Unique identifier for this connection (short, used for tool names)
	Type     string `json:"type"` // mysql or postgres
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`

	// Display metadata (for MCP client context)
	DisplayName string   `json:"display_name,omitempty"` // Full descriptive name (e.g., "Transaction Service Production Database")
	Project     string   `json:"project,omitempty"`      // Project identifier (e.g., "transaction-service")
	Environment string   `json:"environment,omitempty"`  // Environment (e.g., "production", "staging", "dev")
	Description string   `json:"description,omitempty"`  // Detailed description
	Tags        []string `json:"tags,omitempty"`         // Tags for categorization

	// PostgreSQL specific options
	SSLMode            string            `json:"ssl_mode,omitempty"`
	SSLCert            string            `json:"ssl_cert,omitempty"`
	SSLKey             string            `json:"ssl_key,omitempty"`
	SSLRootCert        string            `json:"ssl_root_cert,omitempty"`
	ApplicationName    string            `json:"application_name,omitempty"`
	ConnectTimeout     int               `json:"connect_timeout,omitempty"`
	QueryTimeout       int               `json:"query_timeout,omitempty"` // in seconds
	TargetSessionAttrs string            `json:"target_session_attrs,omitempty"`
	Options            map[string]string `json:"options,omitempty"`

	// Connection pool settings
	MaxOpenConns    int `json:"max_open_conns,omitempty"`
	MaxIdleConns    int `json:"max_idle_conns,omitempty"`
	ConnMaxLifetime int `json:"conn_max_lifetime_seconds,omitempty"`  // in seconds
	ConnMaxIdleTime int `json:"conn_max_idle_time_seconds,omitempty"` // in seconds
}

// MultiDBConfig represents the configuration for multiple database connections
type MultiDBConfig struct {
	Connections []DatabaseConnectionConfig `json:"connections"`
}

// Manager manages multiple database connections
type Manager struct {
	mu          sync.RWMutex
	connections map[string]Database
	configs     map[string]DatabaseConnectionConfig
}

// GetMetadata returns the metadata for a database connection
func (m *Manager) GetMetadata(id string) (DatabaseConnectionConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, ok := m.configs[id]
	return config, ok
}

// NewDBManager creates a new database manager
func NewDBManager() *Manager {
	return &Manager{
		connections: make(map[string]Database),
		configs:     make(map[string]DatabaseConnectionConfig),
	}
}

// LoadConfig loads database configurations from JSON
func (m *Manager) LoadConfig(configJSON []byte) error {
	var config MultiDBConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate and store configurations
	for _, conn := range config.Connections {
		if conn.ID == "" {
			return fmt.Errorf("database connection ID cannot be empty")
		}
		if conn.Type != "mysql" && conn.Type != "postgres" {
			return fmt.Errorf("unsupported database type for connection %s: %s", conn.ID, conn.Type)
		}
		m.configs[conn.ID] = conn
	}

	return nil
}

// Connect establishes connections to all configured databases
// Returns error only if NO databases could be connected
func (m *Manager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var failedConnections []string
	successCount := 0

	// Connect to each database
	for id, cfg := range m.configs {
		// Skip if already connected
		if _, exists := m.connections[id]; exists {
			successCount++
			continue
		}

		// Create database configuration
		dbConfig := Config{
			Type:     cfg.Type,
			Host:     cfg.Host,
			Port:     cfg.Port,
			User:     cfg.User,
			Password: cfg.Password,
			Name:     cfg.Name,
		}

		// Set PostgreSQL-specific options if this is a PostgreSQL database
		if cfg.Type == "postgres" {
			dbConfig.SSLMode = PostgresSSLMode(cfg.SSLMode)
			dbConfig.SSLCert = cfg.SSLCert
			dbConfig.SSLKey = cfg.SSLKey
			dbConfig.SSLRootCert = cfg.SSLRootCert
			dbConfig.ApplicationName = cfg.ApplicationName
			dbConfig.ConnectTimeout = cfg.ConnectTimeout
			dbConfig.QueryTimeout = cfg.QueryTimeout
			dbConfig.TargetSessionAttrs = cfg.TargetSessionAttrs
			dbConfig.Options = cfg.Options
		} else if cfg.Type == "mysql" {
			// Set MySQL-specific options
			dbConfig.ConnectTimeout = cfg.ConnectTimeout
			dbConfig.QueryTimeout = cfg.QueryTimeout
		}

		// Connection pool settings
		if cfg.MaxOpenConns > 0 {
			dbConfig.MaxOpenConns = cfg.MaxOpenConns
		}
		if cfg.MaxIdleConns > 0 {
			dbConfig.MaxIdleConns = cfg.MaxIdleConns
		}
		if cfg.ConnMaxLifetime > 0 {
			dbConfig.ConnMaxLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Second
		}
		if cfg.ConnMaxIdleTime > 0 {
			dbConfig.ConnMaxIdleTime = time.Duration(cfg.ConnMaxIdleTime) * time.Second
		}

		// Create and connect to database
		db, err := NewDatabase(dbConfig)
		if err != nil {
			logger.Warn("Failed to create database instance for %s: %v", id, err)
			failedConnections = append(failedConnections, fmt.Sprintf("%s (create failed)", id))
			continue
		}

		if err := db.Connect(); err != nil {
			logger.Warn("Failed to connect to database %s: %v", id, err)
			failedConnections = append(failedConnections, fmt.Sprintf("%s (connection failed)", id))
			continue
		}

		// Store connected database
		m.connections[id] = db
		successCount++
		logger.Info("Connected to database %s (%s at %s:%d/%s)", id, cfg.Type, cfg.Host, cfg.Port, cfg.Name)
	}

	// Log summary
	if len(failedConnections) > 0 {
		logger.Warn("Failed to connect to %d database(s): %v", len(failedConnections), failedConnections)
	}

	logger.Info("Successfully connected to %d out of %d configured databases", successCount, len(m.configs))

	// Only return error if NO databases connected
	if successCount == 0 {
		return fmt.Errorf("failed to connect to any databases: all %d connection attempts failed", len(m.configs))
	}

	return nil
}

// GetDatabase retrieves a database connection by ID
func (m *Manager) GetDatabase(id string) (Database, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if the database exists
	db, exists := m.connections[id]
	if !exists {
		return nil, fmt.Errorf("database connection %s not found", id)
	}

	return db, nil
}

// GetDatabaseType returns the type of a database by its ID
func (m *Manager) GetDatabaseType(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if the database configuration exists
	cfg, exists := m.configs[id]
	if !exists {
		return "", fmt.Errorf("database configuration %s not found", id)
	}

	return cfg.Type, nil
}

// CloseAll closes all database connections
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error

	// Close each database connection
	for id, db := range m.connections {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database %s: %v", id, err)
			if firstErr == nil {
				firstErr = err
			}
		}
		delete(m.connections, id)
	}

	return firstErr
}

// Close closes a specific database connection
func (m *Manager) Close(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the database exists
	db, exists := m.connections[id]
	if !exists {
		return fmt.Errorf("database connection %s not found", id)
	}

	// Close the connection
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database %s: %w", id, err)
	}

	// Remove from connections map
	delete(m.connections, id)

	return nil
}

// ListDatabases returns a list of all configured databases
func (m *Manager) ListDatabases() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.configs))
	for id := range m.configs {
		ids = append(ids, id)
	}

	return ids
}

// GetConnectedDatabases returns a list of all connected databases
func (m *Manager) GetConnectedDatabases() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.connections))
	for id := range m.connections {
		ids = append(ids, id)
	}

	return ids
}

// GetDatabaseConfig returns the configuration for a specific database
func (m *Manager) GetDatabaseConfig(id string) (DatabaseConnectionConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, exists := m.configs[id]
	if !exists {
		return DatabaseConnectionConfig{}, fmt.Errorf("database configuration %s not found", id)
	}

	return cfg, nil
}
