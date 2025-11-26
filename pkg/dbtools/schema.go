package dbtools

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FreePeak/infra-mcp-server/pkg/db"
	"github.com/FreePeak/infra-mcp-server/pkg/logger"
	"github.com/FreePeak/infra-mcp-server/pkg/tools"
)

// DatabaseStrategy defines the interface for database-specific query strategies
type DatabaseStrategy interface {
	GetTablesQueries() []queryWithArgs
	GetColumnsQueries(table string) []queryWithArgs
	GetRelationshipsQueries(table string) []queryWithArgs
	GetPrimaryKeysQueries(table string) []queryWithArgs
	GetIndexesQueries(table string) []queryWithArgs
	GetEnumValuesQueries() []queryWithArgs
	GetUniqueConstraintsQueries(table string) []queryWithArgs
	GetTableStatsQueries(table string) []queryWithArgs
}

// NewDatabaseStrategy creates the appropriate strategy for the given database type
func NewDatabaseStrategy(driverName string) DatabaseStrategy {
	switch driverName {
	case "postgres":
		return &PostgresStrategy{}
	case "mysql":
		return &MySQLStrategy{}
	default:
		logger.Warn("Unknown database driver: %s, will use generic strategy", driverName)
		return &GenericStrategy{}
	}
}

// PostgresStrategy implements DatabaseStrategy for PostgreSQL
type PostgresStrategy struct{}

// GetTablesQueries returns queries for retrieving tables in PostgreSQL
func (s *PostgresStrategy) GetTablesQueries() []queryWithArgs {
	return []queryWithArgs{
		// Primary: pg_catalog approach
		{query: "SELECT tablename as table_name FROM pg_catalog.pg_tables WHERE schemaname = 'public'"},
		// Secondary: information_schema approach
		{query: "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"},
		// Tertiary: pg_class approach
		{query: "SELECT relname as table_name FROM pg_catalog.pg_class WHERE relkind = 'r' AND relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = 'public')"},
	}
}

// GetColumnsQueries returns queries for retrieving columns in PostgreSQL
func (s *PostgresStrategy) GetColumnsQueries(table string) []queryWithArgs {
	return []queryWithArgs{
		// Primary: information_schema approach for PostgreSQL with UDT name
		{
			query: `
				SELECT 
					column_name, 
					data_type,
					udt_name,
					CASE WHEN is_nullable = 'YES' THEN 'YES' ELSE 'NO' END as is_nullable,
					column_default
				FROM information_schema.columns 
				WHERE table_name = $1 AND table_schema = 'public'
				ORDER BY ordinal_position
			`,
			args: []interface{}{table},
		},
		// Secondary: pg_catalog approach for PostgreSQL
		{
			query: `
				SELECT 
					a.attname as column_name, 
					pg_catalog.format_type(a.atttypid, a.atttypmod) as data_type,
					t.typname as udt_name,
					CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END as is_nullable,
					pg_catalog.pg_get_expr(d.adbin, d.adrelid) as column_default
				FROM pg_catalog.pg_attribute a
				LEFT JOIN pg_catalog.pg_attrdef d ON (a.attrelid = d.adrelid AND a.attnum = d.adnum)
				LEFT JOIN pg_catalog.pg_type t ON a.atttypid = t.oid
				WHERE a.attrelid = (SELECT oid FROM pg_catalog.pg_class WHERE relname = $1 AND relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = 'public'))
				AND a.attnum > 0 AND NOT a.attisdropped
				ORDER BY a.attnum
			`,
			args: []interface{}{table},
		},
	}
}

// GetRelationshipsQueries returns queries for retrieving relationships in PostgreSQL
func (s *PostgresStrategy) GetRelationshipsQueries(table string) []queryWithArgs {
	baseQueries := []queryWithArgs{
		// Primary: Standard information_schema approach for PostgreSQL
		{
			query: `
				SELECT
					tc.table_schema,
					tc.constraint_name,
					tc.table_name,
					kcu.column_name,
					ccu.table_schema AS foreign_table_schema,
					ccu.table_name AS foreign_table_name,
					ccu.column_name AS foreign_column_name
				FROM information_schema.table_constraints AS tc
				JOIN information_schema.key_column_usage AS kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				JOIN information_schema.constraint_column_usage AS ccu
					ON ccu.constraint_name = tc.constraint_name
					AND ccu.table_schema = tc.table_schema
				WHERE tc.constraint_type = 'FOREIGN KEY'
					AND tc.table_schema = 'public'
			`,
			args: []interface{}{},
		},
		// Alternate: Using pg_catalog for older PostgreSQL versions
		{
			query: `
				SELECT
					ns.nspname AS table_schema,
					c.conname AS constraint_name,
					cl.relname AS table_name,
					att.attname AS column_name,
					ns2.nspname AS foreign_table_schema,
					cl2.relname AS foreign_table_name,
					att2.attname AS foreign_column_name
				FROM pg_constraint c
				JOIN pg_class cl ON c.conrelid = cl.oid
				JOIN pg_attribute att ON att.attrelid = cl.oid AND att.attnum = ANY(c.conkey)
				JOIN pg_namespace ns ON ns.oid = cl.relnamespace
				JOIN pg_class cl2 ON c.confrelid = cl2.oid
				JOIN pg_attribute att2 ON att2.attrelid = cl2.oid AND att2.attnum = ANY(c.confkey)
				JOIN pg_namespace ns2 ON ns2.oid = cl2.relnamespace
				WHERE c.contype = 'f'
				AND ns.nspname = 'public'
			`,
			args: []interface{}{},
		},
	}

	if table == "" {
		return baseQueries
	}

	queries := make([]queryWithArgs, len(baseQueries))

	// Add table filter
	queries[0] = queryWithArgs{
		query: baseQueries[0].query + " AND (tc.table_name = $1 OR ccu.table_name = $1)",
		args:  []interface{}{table},
	}

	queries[1] = queryWithArgs{
		query: baseQueries[1].query + " AND (cl.relname = $1 OR cl2.relname = $1)",
		args:  []interface{}{table},
	}

	return queries
}

// GetPrimaryKeysQueries returns queries for retrieving primary keys in PostgreSQL
func (s *PostgresStrategy) GetPrimaryKeysQueries(table string) []queryWithArgs {
	if table == "" {
		return []queryWithArgs{{
			query: `
				SELECT 
					tc.table_name,
					kcu.column_name,
					tc.constraint_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type = 'PRIMARY KEY' 
					AND tc.table_schema = 'public'
				ORDER BY tc.table_name, kcu.ordinal_position
			`,
			args: []interface{}{},
		}}
	}
	
	return []queryWithArgs{
		{
			query: `
				SELECT 
					tc.table_name,
					kcu.column_name,
					tc.constraint_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type = 'PRIMARY KEY' 
					AND tc.table_schema = 'public'
					AND tc.table_name = $1
				ORDER BY kcu.ordinal_position
			`,
			args:  []interface{}{table},
		},
	}
}

// GetIndexesQueries returns queries for retrieving indexes in PostgreSQL
func (s *PostgresStrategy) GetIndexesQueries(table string) []queryWithArgs {
	if table == "" {
		return []queryWithArgs{{
			query: `
				SELECT 
					tablename,
					indexname,
					indexdef
				FROM pg_indexes
				WHERE schemaname = 'public'
				ORDER BY tablename, indexname
			`,
			args: []interface{}{},
		}}
	}
	
	return []queryWithArgs{
		{
			query: `
				SELECT 
					tablename,
					indexname,
					indexdef
				FROM pg_indexes
				WHERE schemaname = 'public'
					AND tablename = $1
				ORDER BY indexname
			`,
			args:  []interface{}{table},
		},
	}
}

// GetEnumValuesQueries returns queries for retrieving ENUM type values in PostgreSQL
func (s *PostgresStrategy) GetEnumValuesQueries() []queryWithArgs {
	return []queryWithArgs{
		{
			query: `
				SELECT 
					t.typname as enum_name,
					n.nspname as schema_name,
					e.enumlabel as enum_value,
					e.enumsortorder as sort_order
				FROM pg_type t
				JOIN pg_enum e ON t.oid = e.enumtypid
				JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
				WHERE n.nspname = 'public'
				ORDER BY t.typname, e.enumsortorder
			`,
			args: []interface{}{},
		},
	}
}

// GetUniqueConstraintsQueries returns queries for retrieving unique constraints in PostgreSQL
func (s *PostgresStrategy) GetUniqueConstraintsQueries(table string) []queryWithArgs {
	if table == "" {
		return []queryWithArgs{{
			query: `
				SELECT 
					tc.table_name,
					tc.constraint_name,
					tc.constraint_type,
					STRING_AGG(kcu.column_name, ', ' ORDER BY kcu.ordinal_position) as column_names
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type IN ('UNIQUE', 'PRIMARY KEY') 
					AND tc.table_schema = 'public'
				GROUP BY tc.table_name, tc.constraint_name, tc.constraint_type
				ORDER BY tc.table_name, tc.constraint_name
			`,
			args: []interface{}{},
		}}
	}
	
	return []queryWithArgs{
		{
			query: `
				SELECT 
					tc.table_name,
					tc.constraint_name,
					tc.constraint_type,
					STRING_AGG(kcu.column_name, ', ' ORDER BY kcu.ordinal_position) as column_names
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type IN ('UNIQUE', 'PRIMARY KEY') 
					AND tc.table_schema = 'public'
					AND tc.table_name = $1
				GROUP BY tc.table_name, tc.constraint_name, tc.constraint_type
				ORDER BY tc.constraint_name
			`,
			args:  []interface{}{table},
		},
	}
}

// GetTableStatsQueries returns queries for retrieving table statistics in PostgreSQL
func (s *PostgresStrategy) GetTableStatsQueries(table string) []queryWithArgs {
	if table == "" {
		return []queryWithArgs{{
			query: `
				SELECT 
					schemaname,
					relname as table_name,
					n_live_tup as row_count_estimate,
					n_dead_tup as dead_tuples,
					last_vacuum,
					last_autovacuum,
					last_analyze,
					last_autoanalyze
				FROM pg_stat_user_tables
				WHERE schemaname = 'public'
				ORDER BY relname
			`,
			args: []interface{}{},
		}}
	}
	
	return []queryWithArgs{
		{
			query: `
				SELECT 
					schemaname,
					relname as table_name,
					n_live_tup as row_count_estimate,
					n_dead_tup as dead_tuples,
					last_vacuum,
					last_autovacuum,
					last_analyze,
					last_autoanalyze
				FROM pg_stat_user_tables
				WHERE schemaname = 'public'
					AND relname = $1
			`,
			args:  []interface{}{table},
		},
	}
}

// MySQLStrategy implements DatabaseStrategy for MySQL
type MySQLStrategy struct{}

// GetTablesQueries returns queries for retrieving tables in MySQL
func (s *MySQLStrategy) GetTablesQueries() []queryWithArgs {
	return []queryWithArgs{
		// Primary: information_schema approach
		{query: "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()"},
		// Secondary: SHOW TABLES approach
		{query: "SHOW TABLES"},
	}
}

// GetColumnsQueries returns queries for retrieving columns in MySQL
func (s *MySQLStrategy) GetColumnsQueries(table string) []queryWithArgs {
	return []queryWithArgs{
		// MySQL query for columns
		{
			query: `
				SELECT column_name, data_type, is_nullable, column_default
				FROM information_schema.columns
				WHERE table_name = ? AND table_schema = DATABASE()
				ORDER BY ordinal_position
			`,
			args: []interface{}{table},
		},
		// Fallback for older MySQL versions
		{
			query: `SHOW COLUMNS FROM ` + table,
			args:  []interface{}{},
		},
	}
}

// GetRelationshipsQueries returns queries for retrieving relationships in MySQL
func (s *MySQLStrategy) GetRelationshipsQueries(table string) []queryWithArgs {
	baseQueries := []queryWithArgs{
		// Primary approach for MySQL
		{
			query: `
				SELECT
					tc.table_schema,
					tc.constraint_name,
					tc.table_name,
					kcu.column_name,
					kcu.referenced_table_schema AS foreign_table_schema,
					kcu.referenced_table_name AS foreign_table_name,
					kcu.referenced_column_name AS foreign_column_name
				FROM information_schema.table_constraints AS tc
				JOIN information_schema.key_column_usage AS kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type = 'FOREIGN KEY'
					AND tc.table_schema = DATABASE()
			`,
			args: []interface{}{},
		},
		// Fallback using simpler query for older MySQL versions
		{
			query: `
				SELECT
					kcu.constraint_schema AS table_schema,
					kcu.constraint_name,
					kcu.table_name,
					kcu.column_name,
					kcu.referenced_table_schema AS foreign_table_schema,
					kcu.referenced_table_name AS foreign_table_name,
					kcu.referenced_column_name AS foreign_column_name
				FROM information_schema.key_column_usage kcu
				WHERE kcu.referenced_table_name IS NOT NULL
					AND kcu.constraint_schema = DATABASE()
			`,
			args: []interface{}{},
		},
	}

	if table == "" {
		return baseQueries
	}

	queries := make([]queryWithArgs, len(baseQueries))

	// Add table filter
	queries[0] = queryWithArgs{
		query: baseQueries[0].query + " AND (tc.table_name = ? OR kcu.referenced_table_name = ?)",
		args:  []interface{}{table, table},
	}

	queries[1] = queryWithArgs{
		query: baseQueries[1].query + " AND (kcu.table_name = ? OR kcu.referenced_table_name = ?)",
		args:  []interface{}{table, table},
	}

	return queries
}

// GetPrimaryKeysQueries returns queries for retrieving primary keys in MySQL
func (s *MySQLStrategy) GetPrimaryKeysQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			tc.table_name,
			kcu.column_name,
			tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY' 
			AND tc.table_schema = DATABASE()
		ORDER BY tc.table_name, kcu.ordinal_position
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " AND tc.table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GetIndexesQueries returns queries for retrieving indexes in MySQL
func (s *MySQLStrategy) GetIndexesQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			table_name,
			index_name,
			GROUP_CONCAT(column_name ORDER BY seq_in_index) as column_names,
			non_unique
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		GROUP BY table_name, index_name, non_unique
		ORDER BY table_name, index_name
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " HAVING table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GetEnumValuesQueries returns queries for retrieving ENUM type values in MySQL
func (s *MySQLStrategy) GetEnumValuesQueries() []queryWithArgs {
	return []queryWithArgs{
		{
			query: `
				SELECT 
					c.table_name,
					c.column_name as enum_name,
					c.column_type as enum_definition
				FROM information_schema.columns c
				WHERE c.table_schema = DATABASE()
					AND c.column_type LIKE 'enum(%'
				ORDER BY c.table_name, c.column_name
			`,
			args: []interface{}{},
		},
	}
}

// GetUniqueConstraintsQueries returns queries for retrieving unique constraints in MySQL
func (s *MySQLStrategy) GetUniqueConstraintsQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			tc.table_name,
			tc.constraint_name,
			tc.constraint_type,
			GROUP_CONCAT(kcu.column_name ORDER BY kcu.ordinal_position) as column_names
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type IN ('UNIQUE', 'PRIMARY KEY') 
			AND tc.table_schema = DATABASE()
		GROUP BY tc.table_name, tc.constraint_name, tc.constraint_type
		ORDER BY tc.table_name, tc.constraint_name
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " HAVING tc.table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GetTableStatsQueries returns queries for retrieving table statistics in MySQL
func (s *MySQLStrategy) GetTableStatsQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			table_schema,
			table_name,
			table_rows as row_count_estimate,
			data_length,
			index_length,
			data_free,
			create_time,
			update_time
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		ORDER BY table_name
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " AND table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GenericStrategy implements DatabaseStrategy for unknown database types
type GenericStrategy struct{}

// GetTablesQueries returns generic queries for retrieving tables
func (s *GenericStrategy) GetTablesQueries() []queryWithArgs {
	return []queryWithArgs{
		{query: "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"},
		{query: "SELECT table_name FROM information_schema.tables"},
		{query: "SHOW TABLES"}, // Last resort
	}
}

// GetColumnsQueries returns generic queries for retrieving columns
func (s *GenericStrategy) GetColumnsQueries(table string) []queryWithArgs {
	return []queryWithArgs{
		// Try PostgreSQL-style query first
		{
			query: `
				SELECT column_name, data_type, is_nullable, column_default
				FROM information_schema.columns
				WHERE table_name = $1
				ORDER BY ordinal_position
			`,
			args: []interface{}{table},
		},
		// Try MySQL-style query
		{
			query: `
				SELECT column_name, data_type, is_nullable, column_default
				FROM information_schema.columns
				WHERE table_name = ?
				ORDER BY ordinal_position
			`,
			args: []interface{}{table},
		},
	}
}

// GetRelationshipsQueries returns generic queries for retrieving relationships
func (s *GenericStrategy) GetRelationshipsQueries(table string) []queryWithArgs {
	pgQuery := queryWithArgs{
		query: `
			SELECT
				tc.table_schema,
				tc.constraint_name,
				tc.table_name,
				kcu.column_name,
				ccu.table_schema AS foreign_table_schema,
				ccu.table_name AS foreign_table_name,
				ccu.column_name AS foreign_column_name
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
		`,
		args: []interface{}{},
	}

	mysqlQuery := queryWithArgs{
		query: `
			SELECT
				kcu.constraint_schema AS table_schema,
				kcu.constraint_name,
				kcu.table_name,
				kcu.column_name,
				kcu.referenced_table_schema AS foreign_table_schema,
				kcu.referenced_table_name AS foreign_table_name,
				kcu.referenced_column_name AS foreign_column_name
			FROM information_schema.key_column_usage kcu
			WHERE kcu.referenced_table_name IS NOT NULL
		`,
		args: []interface{}{},
	}

	if table != "" {
		pgQuery.query += " AND (tc.table_name = $1 OR ccu.table_name = $1)"
		pgQuery.args = append(pgQuery.args, table)

		mysqlQuery.query += " AND (kcu.table_name = ? OR kcu.referenced_table_name = ?)"
		mysqlQuery.args = append(mysqlQuery.args, table, table)
	}

	return []queryWithArgs{pgQuery, mysqlQuery}
}

// GetPrimaryKeysQueries returns queries for retrieving primary keys (generic)
func (s *GenericStrategy) GetPrimaryKeysQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			tc.table_name,
			kcu.column_name,
			tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
		ORDER BY tc.table_name, kcu.ordinal_position
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " AND tc.table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GetIndexesQueries returns queries for retrieving indexes (generic)
func (s *GenericStrategy) GetIndexesQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			table_name,
			index_name,
			column_name
		FROM information_schema.statistics
		ORDER BY table_name, index_name
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " WHERE table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GetEnumValuesQueries returns queries for retrieving ENUM type values (generic)
func (s *GenericStrategy) GetEnumValuesQueries() []queryWithArgs {
	// Try PostgreSQL first, then MySQL
	return []queryWithArgs{
		{
			query: `
				SELECT 
					t.typname as enum_name,
					n.nspname as schema_name,
					e.enumlabel as enum_value,
					e.enumsortorder as sort_order
				FROM pg_type t
				JOIN pg_enum e ON t.oid = e.enumtypid
				JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
				WHERE n.nspname = 'public'
				ORDER BY t.typname, e.enumsortorder
			`,
			args: []interface{}{},
		},
		{
			query: `
				SELECT 
					c.table_name,
					c.column_name as enum_name,
					c.column_type as enum_definition
				FROM information_schema.columns c
				WHERE c.table_schema = DATABASE()
					AND c.column_type LIKE 'enum(%'
				ORDER BY c.table_name, c.column_name
			`,
			args: []interface{}{},
		},
	}
}

// GetUniqueConstraintsQueries returns queries for retrieving unique constraints (generic)
func (s *GenericStrategy) GetUniqueConstraintsQueries(table string) []queryWithArgs {
	baseQuery := `
		SELECT 
			tc.table_name,
			tc.constraint_name,
			tc.constraint_type,
			kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type IN ('UNIQUE', 'PRIMARY KEY')
		ORDER BY tc.table_name, tc.constraint_name, kcu.ordinal_position
	`
	
	if table == "" {
		return []queryWithArgs{{query: baseQuery, args: []interface{}{}}}
	}
	
	return []queryWithArgs{
		{
			query: baseQuery + " AND tc.table_name = ?",
			args:  []interface{}{table},
		},
	}
}

// GetTableStatsQueries returns queries for retrieving table statistics (generic)
func (s *GenericStrategy) GetTableStatsQueries(table string) []queryWithArgs {
	// Try PostgreSQL first
	pgQuery := queryWithArgs{
		query: `
			SELECT 
				schemaname,
				relname as table_name,
				n_live_tup as row_count_estimate,
				n_dead_tup as dead_tuples,
				last_vacuum,
				last_autovacuum,
				last_analyze,
				last_autoanalyze
			FROM pg_stat_user_tables
			WHERE schemaname = 'public'
		`,
		args: []interface{}{},
	}
	
	// MySQL fallback
	mysqlQuery := queryWithArgs{
		query: `
			SELECT 
				table_schema,
				table_name,
				table_rows as row_count_estimate,
				data_length,
				index_length,
				data_free,
				create_time,
				update_time
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
		`,
		args: []interface{}{},
	}
	
	if table != "" {
		pgQuery.query += " AND relname = $1"
		pgQuery.args = append(pgQuery.args, table)
		
		mysqlQuery.query += " AND table_name = ?"
		mysqlQuery.args = append(mysqlQuery.args, table)
	} else {
		pgQuery.query += " ORDER BY relname"
		mysqlQuery.query += " ORDER BY table_name"
	}
	
	return []queryWithArgs{pgQuery, mysqlQuery}
}

// createSchemaExplorerTool creates a tool for exploring database schema
func createSchemaExplorerTool() *tools.Tool {
	return &tools.Tool{
		Name:        "dbSchema",
		Description: "Auto-discover database structure and relationships",
		Category:    "database",
		InputSchema: tools.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"component": map[string]interface{}{
					"type":        "string",
					"description": "Schema component to explore (tables, columns, relationships, or full)",
					"enum":        []string{"tables", "columns", "relationships", "full"},
				},
				"table": map[string]interface{}{
					"type":        "string",
					"description": "Table name to explore (optional, leave empty for all tables)",
				},
				"timeout": map[string]interface{}{
					"type":        "integer",
					"description": "Query timeout in milliseconds (default: 10000)",
				},
				"database": map[string]interface{}{
					"type":        "string",
					"description": "Database ID to use (optional if only one database is configured)",
				},
			},
			Required: []string{"component", "database"},
		},
		Handler: handleSchemaExplorer,
	}
}

// handleSchemaExplorer handles the schema explorer tool execution
func handleSchemaExplorer(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Check if database manager is initialized
	if dbManager == nil {
		return nil, fmt.Errorf("database manager not initialized")
	}

	// Extract parameters
	component, ok := getStringParam(params, "component")
	if !ok {
		return nil, fmt.Errorf("component parameter is required")
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

	// Extract table parameter (optional depending on component)
	table, _ := getStringParam(params, "table")

	// Extract timeout
	timeout := 10000 // Default timeout: 10 seconds
	if timeoutParam, ok := getIntParam(params, "timeout"); ok {
		timeout = timeoutParam
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Use actual database queries based on component type
	switch component {
	case "tables":
		return getTables(timeoutCtx, db)
	case "columns":
		if table == "" {
			return nil, fmt.Errorf("table parameter is required for columns component")
		}
		return getColumns(timeoutCtx, db, table)
	case "relationships":
		return getRelationships(timeoutCtx, db, table)
	case "full":
		return getFullSchema(timeoutCtx, db)
	default:
		return nil, fmt.Errorf("invalid component: %s", component)
	}
}

// executeWithFallbacks executes a series of database queries with fallbacks
// Returns the first successful result or the last error encountered
type queryWithArgs struct {
	query string
	args  []interface{}
}

func executeWithFallbacks(ctx context.Context, db db.Database, queries []queryWithArgs, operationName string) (*sql.Rows, error) {
	var lastErr error

	for i, q := range queries {
		rows, err := db.Query(ctx, q.query, q.args...)
		if err == nil {
			return rows, nil
		}

		lastErr = err
		logger.Warn("%s fallback query %d failed: %v - Error: %v", operationName, i+1, q.query, err)
	}

	// All queries failed, return the last error
	return nil, fmt.Errorf("%s failed after trying %d fallback queries: %w", operationName, len(queries), lastErr)
}

// getTables retrieves the list of tables in the database
func getTables(ctx context.Context, db db.Database) (interface{}, error) {
	// Get database type from connected database
	driverName := db.DriverName()
	dbType := driverName

	// Create the appropriate strategy
	strategy := NewDatabaseStrategy(driverName)

	// Get queries from strategy
	queries := strategy.GetTablesQueries()

	// Execute queries with fallbacks
	rows, err := executeWithFallbacks(ctx, db, queries, "getTables")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	// Convert rows to maps
	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process tables: %w", err)
	}

	return map[string]interface{}{
		"tables": results,
		"dbType": dbType,
	}, nil
}

// getColumns retrieves the columns for a specific table
func getColumns(ctx context.Context, db db.Database, table string) (interface{}, error) {
	// Get database type from connected database
	driverName := db.DriverName()
	dbType := driverName

	// Create the appropriate strategy
	strategy := NewDatabaseStrategy(driverName)

	// Get queries from strategy
	queries := strategy.GetColumnsQueries(table)

	// Execute queries with fallbacks
	rows, err := executeWithFallbacks(ctx, db, queries, "getColumns["+table+"]")
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %w", table, err)
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	// Convert rows to maps
	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process columns: %w", err)
	}

	return map[string]interface{}{
		"table":   table,
		"columns": results,
		"dbType":  dbType,
	}, nil
}

// getRelationships retrieves the relationships for a table or all tables
func getRelationships(ctx context.Context, db db.Database, table string) (interface{}, error) {
	// Get database type from connected database
	driverName := db.DriverName()
	dbType := driverName

	// Create the appropriate strategy
	strategy := NewDatabaseStrategy(driverName)

	// Get queries from strategy
	queries := strategy.GetRelationshipsQueries(table)

	// Execute queries with fallbacks
	rows, err := executeWithFallbacks(ctx, db, queries, "getRelationships")
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships for table %s: %w", table, err)
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	// Convert rows to maps
	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process relationships: %w", err)
	}

	return map[string]interface{}{
		"relationships": results,
		"dbType":        dbType,
		"table":         table,
	}, nil
}

// getPrimaryKeys retrieves the primary keys for a table or all tables
func getPrimaryKeys(ctx context.Context, db db.Database, table string) (interface{}, error) {
	// Get database type from connected database
	driverName := db.DriverName()
	dbType := driverName

	// Create the appropriate strategy
	strategy := NewDatabaseStrategy(driverName)

	// Get queries from strategy
	queries := strategy.GetPrimaryKeysQueries(table)

	// Execute queries with fallbacks
	rows, err := executeWithFallbacks(ctx, db, queries, "getPrimaryKeys")
	if err != nil {
		return nil, fmt.Errorf("failed to get primary keys for table %s: %w", table, err)
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	// Convert rows to maps
	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process primary keys: %w", err)
	}

	return map[string]interface{}{
		"primary_keys": results,
		"dbType":       dbType,
		"table":        table,
	}, nil
}

// getIndexes retrieves the indexes for a table or all tables
func getIndexes(ctx context.Context, db db.Database, table string) (interface{}, error) {
	// Get database type from connected database
	driverName := db.DriverName()
	dbType := driverName

	// Create the appropriate strategy
	strategy := NewDatabaseStrategy(driverName)

	// Get queries from strategy
	queries := strategy.GetIndexesQueries(table)

	// Execute queries with fallbacks
	rows, err := executeWithFallbacks(ctx, db, queries, "getIndexes")
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes for table %s: %w", table, err)
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	// Convert rows to maps
	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process indexes: %w", err)
	}

	return map[string]interface{}{
		"indexes": results,
		"dbType":  dbType,
		"table":   table,
	}, nil
}

// getEnumValues retrieves ENUM type definitions and their values
func getEnumValues(ctx context.Context, db db.Database) (interface{}, error) {
	driverName := db.DriverName()
	dbType := driverName

	strategy := NewDatabaseStrategy(driverName)
	queries := strategy.GetEnumValuesQueries()

	rows, err := executeWithFallbacks(ctx, db, queries, "getEnumValues")
	if err != nil {
		// Don't fail if enums aren't supported, just return empty result
		logger.Warn("Failed to get enum values (may not be supported): %v", err)
		return map[string]interface{}{
			"enums":  []map[string]interface{}{},
			"dbType": dbType,
		}, nil
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process enum values: %w", err)
	}

	return map[string]interface{}{
		"enums":  results,
		"dbType": dbType,
	}, nil
}

// getUniqueConstraints retrieves unique constraints for a table or all tables
func getUniqueConstraints(ctx context.Context, db db.Database, table string) (interface{}, error) {
	driverName := db.DriverName()
	dbType := driverName

	strategy := NewDatabaseStrategy(driverName)
	queries := strategy.GetUniqueConstraintsQueries(table)

	rows, err := executeWithFallbacks(ctx, db, queries, "getUniqueConstraints")
	if err != nil {
		return nil, fmt.Errorf("failed to get unique constraints: %w", err)
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process unique constraints: %w", err)
	}

	return map[string]interface{}{
		"unique_constraints": results,
		"dbType":             dbType,
		"table":              table,
	}, nil
}

// getTableStats retrieves table statistics (row counts, etc.)
func getTableStats(ctx context.Context, db db.Database, table string) (interface{}, error) {
	driverName := db.DriverName()
	dbType := driverName

	strategy := NewDatabaseStrategy(driverName)
	queries := strategy.GetTableStatsQueries(table)

	rows, err := executeWithFallbacks(ctx, db, queries, "getTableStats")
	if err != nil {
		// Don't fail if stats aren't available
		logger.Warn("Failed to get table stats: %v", err)
		return map[string]interface{}{
			"stats":  []map[string]interface{}{},
			"dbType": dbType,
			"table":  table,
		}, nil
	}

	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				logger.Error("error closing rows: %v", err)
			}
		}
	}()

	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to process table stats: %w", err)
	}

	return map[string]interface{}{
		"stats":  results,
		"dbType": dbType,
		"table":  table,
	}, nil
}

// safeGetMap safely gets a map from an interface value
func safeGetMap(obj interface{}) (map[string]interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil value cannot be converted to map")
	}

	mapVal, ok := obj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value is not a map[string]interface{}: %T", obj)
	}

	return mapVal, nil
}

// safeGetString safely gets a string from a map key
func safeGetString(m map[string]interface{}, key string) (string, error) {
	val, ok := m[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in map", key)
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("value for key %q is not a string: %T", key, val)
	}

	return strVal, nil
}

// getFullSchema retrieves the complete database schema
func getFullSchema(ctx context.Context, db db.Database) (interface{}, error) {
	tablesResult, err := getTables(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	tablesMap, err := safeGetMap(tablesResult)
	if err != nil {
		return nil, fmt.Errorf("invalid tables result: %w", err)
	}

	tablesSlice, ok := tablesMap["tables"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tables data format")
	}

	// Get ENUM values for all types
	enumsResult, enumsErr := getEnumValues(ctx, db)
	var enumValues []map[string]interface{}
	var enumsByType map[string][]string
	if enumsErr != nil {
		logger.Warn("Failed to get enum values: %v", enumsErr)
		enumValues = []map[string]interface{}{}
		enumsByType = make(map[string][]string)
	} else {
		enumsMap, _ := safeGetMap(enumsResult)
		if enums, ok := enumsMap["enums"].([]map[string]interface{}); ok {
			enumValues = enums
		}
		
		// Organize enum values by type name for easy lookup
		enumsByType = make(map[string][]string)
		for _, enum := range enumValues {
			if enumName, ok := enum["enum_name"].(string); ok {
				if enumValue, ok := enum["enum_value"].(string); ok {
					enumsByType[enumName] = append(enumsByType[enumName], enumValue)
				}
			}
		}
	}

	// Get table statistics for all tables
	statsResult, statsErr := getTableStats(ctx, db, "")
	var statsByTable map[string]map[string]interface{}
	if statsErr != nil {
		logger.Warn("Failed to get table stats: %v", statsErr)
		statsByTable = make(map[string]map[string]interface{})
	} else {
		statsMap, _ := safeGetMap(statsResult)
		if stats, ok := statsMap["stats"].([]map[string]interface{}); ok {
			statsByTable = make(map[string]map[string]interface{})
			for _, stat := range stats {
				if tableName, ok := stat["table_name"].(string); ok {
					statsByTable[tableName] = stat
				}
			}
		}
	}

	// For each table, get detailed information
	detailedSchema := make(map[string]interface{})
	for _, tableInfo := range tablesSlice {
		tableName, err := safeGetString(tableInfo, "table_name")
		if err != nil {
			return nil, fmt.Errorf("invalid table info: %w", err)
		}

		// Get columns
		columnsResult, columnsErr := getColumns(ctx, db, tableName)
		if columnsErr != nil {
			logger.Warn("Failed to get columns for table %s: %v", tableName, columnsErr)
			continue
		}
		
		columnsMap, _ := safeGetMap(columnsResult)
		
		// Enhance columns with enum values
		if columns, ok := columnsMap["columns"].([]map[string]interface{}); ok {
			for i, column := range columns {
				dataType, _ := column["data_type"].(string)
				udtName, hasUdtName := column["udt_name"].(string)
				
				// For USER-DEFINED types, use the udt_name to look up enum values
				if dataType == "USER-DEFINED" && hasUdtName {
					if enumVals, exists := enumsByType[udtName]; exists {
						columns[i]["enum_values"] = enumVals
						columns[i]["enum_type"] = udtName
					}
				} else {
					// For other types, check by data_type
					if enumVals, exists := enumsByType[dataType]; exists {
						columns[i]["enum_values"] = enumVals
					}
				}
			}
		}
		
		// Get primary keys for this table
		primaryKeysResult, pkErr := getPrimaryKeys(ctx, db, tableName)
		var primaryKeys []map[string]interface{}
		if pkErr != nil {
			logger.Warn("Failed to get primary keys for table %s: %v", tableName, pkErr)
			primaryKeys = []map[string]interface{}{}
		} else {
			pkMap, _ := safeGetMap(primaryKeysResult)
			if pks, ok := pkMap["primary_keys"].([]map[string]interface{}); ok {
				primaryKeys = pks
			}
		}
		
		// Get indexes for this table
		indexesResult, idxErr := getIndexes(ctx, db, tableName)
		var indexes []map[string]interface{}
		if idxErr != nil {
			logger.Warn("Failed to get indexes for table %s: %v", tableName, idxErr)
			indexes = []map[string]interface{}{}
		} else {
			idxMap, _ := safeGetMap(indexesResult)
			if idxs, ok := idxMap["indexes"].([]map[string]interface{}); ok {
				indexes = idxs
			}
		}
		
		// Get unique constraints for this table
		uniqueConstraintsResult, ucErr := getUniqueConstraints(ctx, db, tableName)
		var uniqueConstraints []map[string]interface{}
		if ucErr != nil {
			logger.Warn("Failed to get unique constraints for table %s: %v", tableName, ucErr)
			uniqueConstraints = []map[string]interface{}{}
		} else {
			ucMap, _ := safeGetMap(uniqueConstraintsResult)
			if ucs, ok := ucMap["unique_constraints"].([]map[string]interface{}); ok {
				uniqueConstraints = ucs
			}
		}

		// Get table statistics
		tableStats := statsByTable[tableName]
		if tableStats == nil {
			tableStats = make(map[string]interface{})
		}

		// Build detailed table schema
		detailedSchema[tableName] = map[string]interface{}{
			"columns":            columnsMap["columns"],
			"primary_keys":       primaryKeys,
			"indexes":            indexes,
			"unique_constraints": uniqueConstraints,
			"statistics":         tableStats,
		}
	}

	// Get all relationships
	relationships, relErr := getRelationships(ctx, db, "")
	var foreignKeys []map[string]interface{}
	if relErr != nil {
		logger.Warn("Failed to get relationships: %v", relErr)
		foreignKeys = []map[string]interface{}{}
	} else {
		relMap, _ := safeGetMap(relationships)
		if fks, ok := relMap["relationships"].([]map[string]interface{}); ok {
			foreignKeys = fks
		}
	}

	// Organize foreign keys by table
	fksByTable := make(map[string][]map[string]interface{})
	for _, fk := range foreignKeys {
		if tableName, ok := fk["table_name"].(string); ok {
			fksByTable[tableName] = append(fksByTable[tableName], fk)
		}
	}
	
	// Add foreign keys to each table's detailed schema
	for tableName, tableSchema := range detailedSchema {
		if schema, ok := tableSchema.(map[string]interface{}); ok {
			schema["foreign_keys"] = fksByTable[tableName]
		}
	}

	return map[string]interface{}{
		"tables":          tablesSlice,
		"detailed_schema": detailedSchema,
		"foreign_keys":    foreignKeys,
		"enum_types":      enumsByType,
		"enum_values":     enumValues,
	}, nil
}
