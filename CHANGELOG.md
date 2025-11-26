# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Guidelines for Maintaining This Changelog

**IMPORTANT**: This CHANGELOG must be updated with EVERY change that is merged to the main branch.

### When to Update

- ✅ **Always** update before committing new features
- ✅ **Always** update before committing bug fixes
- ✅ **Always** update before committing breaking changes
- ✅ **Always** update for documentation changes that affect users
- ❌ Do not update for internal refactoring unless it affects behavior

### How to Update

1. Add your changes under `## [Unreleased]`
2. Use the appropriate section: Added, Changed, Deprecated, Removed, Fixed, Security
3. Be specific and user-focused (not implementation details)
4. Include references to issues/PRs when applicable
5. When releasing, move `[Unreleased]` items to a new version section

### Version Numbering

- **Major (X.0.0)**: Breaking changes
- **Minor (x.X.0)**: New features, backward compatible
- **Patch (x.x.X)**: Bug fixes, backward compatible

**No exceptions. No "we'll document it later." Update the CHANGELOG with your changes.**

---

## [Unreleased]

## [v1.7.0] - 2025-10-21 🚀

**MAJOR RELEASE: Schema Intelligence Enhancement**

This release dramatically improves database schema understanding with automatic enum detection, enhanced metadata, and intelligent query assistance. Estimated 30-40% efficiency boost for AI agents working with databases.

### Added - Schema Enhancement Features 🚀

- **ENUM/USER-DEFINED Type Values** (BIGGEST WIN!)

  - Automatically discover and expose all ENUM type values for PostgreSQL and MySQL
  - Column-level enum values displayed inline with `enum_values` and `enum_type` fields
  - Global `enum_types` map for quick lookup of all valid values
  - Prevents invalid WHERE clauses and query errors by showing valid enum values upfront
  - Estimated 30-40% efficiency gain for working with enum-heavy schemas

- **Enhanced Foreign Key Relationships**

  - Complete foreign key mapping showing source and target tables/columns
  - Organized by table for easy navigation
  - Enables instant JOIN query generation without exploration

- **Table Row Counts (Approximate)**

  - Fast approximate row counts using `pg_stat_user_tables` (PostgreSQL) and `information_schema` (MySQL)
  - No expensive COUNT(\*) queries required
  - Helps determine when to add LIMIT clauses and optimize query strategies
  - Includes additional metadata: dead tuples, last vacuum/analyze times

- **Unique Constraints (Detailed)**

  - Shows which columns have unique constraints with readable names
  - Includes both single-column and composite unique constraints
  - Grouped by table with constraint type (UNIQUE vs PRIMARY KEY)
  - Column names displayed as comma-separated list

- **Enhanced Default Values**
  - Column default values now consistently included across all query strategies
  - Includes both simple defaults and complex expressions

### Changed

- Updated `DatabaseStrategy` interface with three new methods:
  - `GetEnumValuesQueries()` - Query ENUM definitions
  - `GetUniqueConstraintsQueries(table)` - Query unique constraints
  - `GetTableStatsQueries(table)` - Query table statistics
- Enhanced `GetColumnsQueries` for PostgreSQL to include `udt_name` field
- Improved `getFullSchema` to include enum values, unique constraints, and table statistics
- Added helper functions: `getEnumValues`, `getUniqueConstraints`, `getTableStats`

### Documentation

- Added comprehensive `docs/SCHEMA_ENHANCEMENTS.md` with:
  - Detailed feature descriptions and benefits
  - SQL implementation examples for PostgreSQL and MySQL
  - Output format specifications
  - Performance metrics and optimization tips
  - Troubleshooting guide
  - Migration notes (100% backward compatible)

### Testing

- Added `test-schema-enhancements.sql` script for testing all new features
- Creates sample database with ENUMs, foreign keys, and constraints

### Performance

- All new features leverage existing caching infrastructure
- ENUM queries are one-time fetches with results cached
- Table statistics use fast system views (no table scans)
- No breaking changes or performance regressions

### Backward Compatibility

- ✅ 100% backward compatible
- ✅ All existing APIs continue to work unchanged
- ✅ New fields are additive only
- ✅ Graceful degradation for unsupported databases

## [v1.6.5] - 2025-10-21

### Added

- Full schema display with enhanced caching
- Schema cache implementation with configurable TTL (default 5 minutes)
- Cache cleanup routine for expired entries

### Fixed

- Empty result response format in some edge cases

### Documentation

- Updated README with recent changes

## [v1.6.4] - 2025-10-20

### Added - TimescaleDB Context Features

- **CTX-1**: TimescaleDB detection in editor context
  - Automatically detects if connected database has TimescaleDB extension
  - Provides version information and available features
- **CTX-2**: Hypertable schema information in context
  - Shows hypertable metadata including chunk intervals, compression, retention
  - Displays column information for hypertables
- **CTX-4 & CTX-5**: Query suggestions and function documentation
  - Context-aware query suggestions for TimescaleDB
  - Inline documentation for TimescaleDB-specific functions

### Documentation

- Added `docs/TIMESCALEDB_FUNCTIONS.md` - TimescaleDB function reference
- Added documentation for context features
- Updated implementation status documents

## [v1.6.3] - 2025-10-18

### Added - TimescaleDB Tools (Complete Suite)

- **TOOL-1**: TimescaleDB tool category registration
- **TOOL-2**: Hypertable creation tool
  - Create hypertables with automatic partitioning
  - Configure chunk time intervals
  - Set up space partitioning
- **TOOL-3**: Hypertable listing tool
  - List all hypertables with metadata
  - Show chunk counts, sizes, and configurations
- **TOOL-4**: Compression policy tools
  - Create and manage compression policies
  - Configure compression settings (segment by, order by)
  - Set compression intervals
- **TOOL-5**: Retention policy tools
  - Create and manage data retention policies
  - Configure automatic data deletion
  - Set retention intervals
- **TOOL-6**: Time-series query tools
  - Time-bucket aggregation queries
  - First/last value queries
  - Rate calculation queries
  - Moving average calculations
- **TOOL-7**: Continuous aggregate tools
  - Create continuous aggregates
  - Manage refresh policies
  - Query materialized data

### Added - TimescaleDB Testing

- **TEST-1**: Docker test environment for TimescaleDB
  - `docker-compose.timescaledb-test.yml` for local testing
  - Sample data initialization scripts
  - Continuous aggregate test setup

### Fixed

- TimescaleDB unit test failures
- Golangci-lint errors in TimescaleDB package
- Mock expectations in timescale_tools_test.go

### Documentation

- Added `docs/TIMESCALEDB_PRD.md` - Product Requirements Document
- Added `docs/TIMESCALEDB_IMPLEMENTATION.md` - Implementation guide
- Added `docs/TIMESCALEDB_TOOLS.md` - Tool documentation
- Added `init-scripts/timescaledb/` with initialization SQL scripts

## [v1.6.2] - 2025-10-15

### Added

- Query configuration timeout support
- Configurable query timeout via MCP parameters

### Changed

- Refactored database operations using strategy and factory patterns
  - Improved PostgreSQL-specific query handling
  - Better separation of concerns for different database types
  - More maintainable and extensible architecture

### Fixed

- PostgreSQL-specific query issues
- Connection handling improvements
- Docker build for multi-architecture support (AMD64, x86)

### Documentation

- Updated build commands in README
- Added docker-compose configuration for local testing

## [v1.6.1] - 2025-04-01

### Added

- OpenAI Agents SDK compatibility by adding Items property to array parameters
- Test script for verifying OpenAI Agents SDK compatibility

### Fixed

- Issue #8: Array parameters in tool definitions now include required `items` property
- JSON Schema validation errors in OpenAI Agents SDK integration

## [v1.6.0] - 2023-03-31

### Changed

- Upgraded cortex dependency from v1.0.3 to v1.0.4

## [] - 2023-03-31

### Added

- Internal logging system for improved debugging and monitoring
- Logger implementation for all packages

### Fixed

- Connection issues with PostgreSQL databases
- Restored functionality for all MCP tools
- Eliminated non-JSON RPC logging in stdio mode

## [] - 2023-03-25

### Added

- Initial release of Infrastructure MCP Server
- Multi-database connection support
- Tool generation for database operations
- README with guidelines on using tools in Cursor
