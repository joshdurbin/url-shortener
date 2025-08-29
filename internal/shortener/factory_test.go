package shortener

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/joshdurbin/url-shortener/db/sqlc"
)

func setupFactoryTestDB(t *testing.T) *sqlc.Queries {
	dbPath := filepath.Join(t.TempDir(), "factory_test.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	
	// Create counters table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS counters (
			key TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create counters table: %v", err)
	}
	
	return sqlc.New(db)
}

func TestNewGenerator(t *testing.T) {
	queries := setupFactoryTestDB(t)

	testCases := []struct {
		name        string
		config      Config
		requiresDB  bool
		expectedType string
		shouldError bool
	}{
		{
			name: "Counter generator",
			config: Config{
				CounterStep: 100,
			},
			requiresDB:   true,
			expectedType: TypeCounter,
			shouldError:  false,
		},
		{
			name: "Counter generator with step 1",
			config: Config{
				CounterStep: 1,
			},
			requiresDB:   true,
			expectedType: TypeCounter,
			shouldError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var db *sqlc.Queries
			if tc.requiresDB {
				db = queries
			}

			generator, err := NewGenerator(tc.config, db)

			if tc.shouldError {
				if err == nil {
					t.Error("Expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewGenerator failed: %v", err)
			}

			if generator == nil {
				t.Fatal("Expected generator, got nil")
			}

			defer generator.Close()

			if generator.Type() != tc.expectedType {
				t.Errorf("Expected generator type %s, got %s", tc.expectedType, generator.Type())
			}
		})
	}

	t.Run("Counter generator without database fails", func(t *testing.T) {
		config := Config{
			CounterStep: 100,
		}

		generator, err := NewGenerator(config, nil)
		if err == nil {
			t.Error("Expected error when creating counter generator without database")
			if generator != nil {
				generator.Close()
			}
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	expectedDefaults := Config{
		CounterStep: 1,
	}

	if config.CounterStep != expectedDefaults.CounterStep {
		t.Errorf("Expected default counter step %d, got %d", expectedDefaults.CounterStep, config.CounterStep)
	}
}