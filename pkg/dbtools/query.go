package dbtools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/FreePeak/infra-mcp-server/pkg/logger"
	"github.com/FreePeak/infra-mcp-server/pkg/tools"
)

// createQueryTool creates a tool for executing database queries
//
//nolint:unused // Retained for future use
func createQueryTool() *tools.Tool {
	return &tools.Tool{
		Name:        "dbQuery",
		Description: "Execute a database query that returns results",
		Category:    "database",
		InputSchema: tools.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "SQL query to execute",
				},
				"params": map[string]interface{}{
					"type":        "array",
					"description": "Parameters for the query (for prepared statements)",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"timeout": map[string]interface{}{
					"type":        "integer",
					"description": "Query timeout in milliseconds (default: 5000)",
				},
				"database": map[string]interface{}{
					"type":        "string",
					"description": "Database ID to use (optional if only one database is configured)",
				},
			},
			Required: []string{"query", "database"},
		},
		Handler: handleQuery,
	}
}

// handleQuery handles the query tool execution
func handleQuery(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Check if database manager is initialized
	if dbManager == nil {
		return nil, fmt.Errorf("database manager not initialized")
	}

	// Extract parameters
	query, ok := getStringParam(params, "query")
	if !ok {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Validate that the query is read-only
	if err := validateReadOnlyQuery(query); err != nil {
		return nil, err
	}

	// Get database ID
	databaseID, ok := getStringParam(params, "database")
	if !ok {
		return nil, fmt.Errorf("database parameter is required")
	}

	// Get database instance
	db, err := dbManager.GetDatabase(databaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Extract timeout
	dbTimeout := db.QueryTimeout() * 1000 // Convert from seconds to milliseconds
	timeout := dbTimeout                  // Default to the database's query timeout
	if timeoutParam, ok := getIntParam(params, "timeout"); ok {
		timeout = timeoutParam
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Extract query parameters
	var queryParams []interface{}
	if paramsArray, ok := getArrayParam(params, "params"); ok {
		queryParams = make([]interface{}, len(paramsArray))
		copy(queryParams, paramsArray)
	}

	// Get the performance analyzer
	analyzer := GetPerformanceAnalyzer()

	// Execute query with performance tracking
	var result interface{}

	result, err = analyzer.TrackQuery(timeoutCtx, query, queryParams, func() (interface{}, error) {
		// Execute query
		rows, innerErr := db.Query(timeoutCtx, query, queryParams...)
		if innerErr != nil {
			return nil, fmt.Errorf("failed to execute query: %w", innerErr)
		}
		defer cleanupRows(rows)

		// Convert rows to maps
		results, innerErr := rowsToMaps(rows)
		if innerErr != nil {
			return nil, fmt.Errorf("failed to process query results: %w", innerErr)
		}

		return map[string]interface{}{
			"results":  results,
			"query":    query,
			"params":   queryParams,
			"rowCount": len(results),
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// containsIgnoreCase checks if a string contains a substring, ignoring case
//
//nolint:unused // Retained for future use
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// validateReadOnlyQuery checks if a query contains only read-only operations
func validateReadOnlyQuery(query string) error {
	// Normalize query to uppercase for checking
	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// List of write operation keywords that should be rejected
	writeKeywords := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"TRUNCATE", "REPLACE", "MERGE", "GRANT", "REVOKE",
		"EXEC", "EXECUTE", "CALL",
	}

	// Check if the query starts with any write operation
	for _, keyword := range writeKeywords {
		if strings.HasPrefix(upperQuery, keyword) {
			return fmt.Errorf("write operations are not allowed in read-only mode: detected %s statement", keyword)
		}
		// Also check for write operations after comments or whitespace
		if strings.Contains(upperQuery, ";"+keyword) || strings.Contains(upperQuery, "; "+keyword) {
			return fmt.Errorf("write operations are not allowed in read-only mode: detected %s statement", keyword)
		}
	}

	// Additional check for common write patterns in the middle of queries
	// This catches cases like "SELECT ... INTO", "WITH ... INSERT", etc.
	dangerousPatterns := []string{
		"INSERT INTO", "UPDATE ", "DELETE FROM", "DROP ", "CREATE ",
		"ALTER ", "TRUNCATE ", "INTO OUTFILE", "INTO DUMPFILE",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperQuery, pattern) {
			return fmt.Errorf("write operations are not allowed in read-only mode: detected '%s' pattern", pattern)
		}
	}

	// Check for SELECT INTO pattern (but allow INTO OUTFILE/DUMPFILE which are already caught)
	if strings.Contains(upperQuery, " INTO ") {
		// This could be SELECT INTO or INSERT INTO
		// INSERT INTO is already checked, so this catches SELECT INTO
		if !strings.Contains(upperQuery, "INSERT INTO") {
			return fmt.Errorf("write operations are not allowed in read-only mode: detected 'INTO' clause")
		}
	}

	return nil
}

// cleanupRows ensures rows are closed properly
func cleanupRows(rows *sql.Rows) {
	if rows != nil {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Error("error closing rows: %v", closeErr)
		}
	}
}
