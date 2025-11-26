package dbtools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FreePeak/infra-mcp-server/internal/logger"
	pkgLogger "github.com/FreePeak/infra-mcp-server/pkg/logger"
)

// TestEnumDetectionLive tests enum detection with a real database connection
// Run with: go test -v -run TestEnumDetectionLive ./pkg/dbtools/
func TestEnumDetectionLive(t *testing.T) {
	// Skip if in CI or automated environment
	if testing.Short() {
		t.Skip("Skipping live database test")
	}

	// Initialize logger first (required by InitDatabase)
	logger.Initialize("error")
	pkgLogger.Initialize("error")

	// Initialize using the bin/config.json
	cfg := &Config{
		ConfigFile: "../../bin/config.json",
	}

	if err := InitDatabase(cfg); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Get the database
	if dbManager == nil {
		t.Fatal("dbManager is nil after InitDatabase")
	}

	db, err := dbManager.GetDatabase("ts_stage")
	if err != nil {
		t.Fatalf("Failed to get database: %v", err)
	}

	ctx := context.Background()

	// Test getEnumValues
	t.Run("GetEnumValues", func(t *testing.T) {
		result, err := getEnumValues(ctx, db)
		if err != nil {
			t.Fatalf("getEnumValues failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Result is not a map: %T", result)
		}

		enums, ok := resultMap["enums"].([]map[string]interface{})
		if !ok {
			t.Fatalf("enums is not []map[string]interface{}: %T", resultMap["enums"])
		}

		t.Logf("Found %d enum values", len(enums))
		
		if len(enums) == 0 {
			t.Error("❌ NO enum values found - getEnumValues returned empty")
		} else {
			t.Logf("✅ Found enum values:")
			for i, enum := range enums {
				if i >= 5 {
					break
				}
				t.Logf("  - %v", enum)
			}
		}
	})

	// Test getFullSchema
	t.Run("GetFullSchema", func(t *testing.T) {
		result, err := getFullSchema(ctx, db)
		if err != nil {
			t.Fatalf("getFullSchema failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Result is not a map: %T", result)
		}

		// Check for enum_types
		t.Run("EnumTypes", func(t *testing.T) {
			enumTypes, ok := resultMap["enum_types"]
			if !ok {
				t.Error("❌ enum_types NOT FOUND in result")
				t.Log("Available keys:")
				for key := range resultMap {
					t.Logf("  - %s", key)
				}
				return
			}

			enumTypesMap, ok := enumTypes.(map[string][]string)
			if !ok {
				t.Errorf("❌ enum_types has wrong type: %T", enumTypes)
				return
			}

			if len(enumTypesMap) == 0 {
				t.Error("❌ enum_types is EMPTY")
			} else {
				t.Logf("✅ Found %d enum types:", len(enumTypesMap))
				for typeName, values := range enumTypesMap {
					t.Logf("  - %s: %v", typeName, values)
				}
			}
		})

		// Check for enum_values
		t.Run("EnumValues", func(t *testing.T) {
			enumValues, ok := resultMap["enum_values"]
			if !ok {
				t.Error("❌ enum_values NOT FOUND in result")
				return
			}

			enumValuesArray, ok := enumValues.([]map[string]interface{})
			if !ok {
				t.Errorf("❌ enum_values has wrong type: %T", enumValues)
				return
			}

			t.Logf("✅ Found %d enum value entries", len(enumValuesArray))
		})

		// Check column-level enum values
		t.Run("ColumnEnumValues", func(t *testing.T) {
			detailedSchema, ok := resultMap["detailed_schema"].(map[string]interface{})
			if !ok {
				t.Error("❌ detailed_schema not found")
				return
			}

			transTable, ok := detailedSchema["transactions"].(map[string]interface{})
			if !ok {
				t.Error("❌ transactions table not found")
				return
			}

			columns, ok := transTable["columns"].([]map[string]interface{})
			if !ok {
				t.Error("❌ columns not found or wrong type")
				return
			}

			t.Logf("Checking %d columns in transactions table", len(columns))
			
			enumCount := 0
			for _, column := range columns {
				colName := column["column_name"]
				dataType := column["data_type"]
				udtName := column["udt_name"]

				if dataType == "USER-DEFINED" {
					if enumValues, hasEnum := column["enum_values"]; hasEnum {
						enumCount++
						t.Logf("  ✅ %v (udt: %v) has enum_values: %v", colName, udtName, enumValues)
					} else {
						t.Logf("  ⚠️  %v (udt: %v) is USER-DEFINED but NO enum_values", colName, udtName)
					}
				}
			}

			if enumCount == 0 {
				t.Error("❌ NO columns have enum_values attached")
			} else {
				t.Logf("✅ Found %d columns with enum values", enumCount)
			}
		})

		// Pretty print a sample for debugging
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		t.Logf("\n\nFull result sample (first 2000 chars):\n%s\n...", string(jsonBytes[:min(2000, len(jsonBytes))]))
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

