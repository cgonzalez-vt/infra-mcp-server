package dbtools

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/FreePeak/infra-mcp-server/pkg/logger"
)

// SchemaCache provides a thread-safe cache for database schema information
type SchemaCache struct {
	mu      sync.RWMutex
	entries map[string]*schemaCacheEntry
	ttl     time.Duration
}

// schemaCacheEntry holds a cached schema with timestamp
type schemaCacheEntry struct {
	schema    interface{}
	timestamp time.Time
}

// Global schema cache instance
var schemaCache *SchemaCache

// InitSchemaCache initializes the schema cache with the configured TTL
func InitSchemaCache() {
	ttl := getSchemaCacheTTL()
	schemaCache = &SchemaCache{
		entries: make(map[string]*schemaCacheEntry),
		ttl:     ttl,
	}
	logger.Info("Schema cache initialized with TTL: %v", ttl)
}

// GetSchemaCache returns the global schema cache instance
func GetSchemaCache() *SchemaCache {
	if schemaCache == nil {
		InitSchemaCache()
	}
	return schemaCache
}

// Get retrieves a cached schema if it exists and hasn't expired
func (c *SchemaCache) Get(dbID string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[dbID]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}

	logger.Debug("Schema cache hit for database: %s", dbID)
	return entry.schema, true
}

// Set stores a schema in the cache
func (c *SchemaCache) Set(dbID string, schema interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[dbID] = &schemaCacheEntry{
		schema:    schema,
		timestamp: time.Now(),
	}
	
	logger.Debug("Schema cached for database: %s", dbID)
}

// Invalidate removes a schema from the cache
func (c *SchemaCache) Invalidate(dbID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, dbID)
	logger.Debug("Schema cache invalidated for database: %s", dbID)
}

// InvalidateAll clears the entire cache
func (c *SchemaCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*schemaCacheEntry)
	logger.Info("Schema cache cleared")
}

// CleanupExpired removes expired entries from the cache
func (c *SchemaCache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for dbID, entry := range c.entries {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.entries, dbID)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		logger.Debug("Cleaned up %d expired schema cache entries", expiredCount)
	}
}

// getSchemaCacheTTL reads the cache TTL from environment variable or returns default
func getSchemaCacheTTL() time.Duration {
	ttlStr := os.Getenv("SCHEMA_CACHE_TTL")
	if ttlStr == "" {
		return 5 * time.Minute // Default: 5 minutes
	}

	ttlSeconds, err := strconv.Atoi(ttlStr)
	if err != nil {
		logger.Warn("Invalid SCHEMA_CACHE_TTL value '%s', using default 5 minutes", ttlStr)
		return 5 * time.Minute
	}

	return time.Duration(ttlSeconds) * time.Second
}

// StartCleanupRoutine starts a background goroutine to periodically clean up expired entries
func (c *SchemaCache) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(c.ttl)
		defer ticker.Stop()

		for range ticker.C {
			c.CleanupExpired()
		}
	}()
}

