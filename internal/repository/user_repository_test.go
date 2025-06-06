package repository

import (
	"assignment/internal/domain/entities"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// runMigrations executes all .up.sql migration files in order
func runMigrations(t *testing.T, db *sql.DB) error {
	migrationsDir := filepath.Join("..", "..", "db", "migrations")

	// Read all files in migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Filter and sort .up.sql files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute each migration file in order
	for _, filename := range migrationFiles {
		migrationPath := filepath.Join(migrationsDir, filename)
		migrationBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			return err
		}

		migrationSQL := string(migrationBytes)
		if _, err := db.Exec(migrationSQL); err != nil {
			t.Logf("Failed to execute migration %s: %v", filename, err)
			return err
		}

		t.Logf("Successfully executed migration: %s", filename)
	}

	return nil
}

func setupTestContainer(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Run all migration files (single source of truth)
	if err := runMigrations(t, db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

func TestUserRepository_CreateFriendship(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		user1   *entities.User
		user2   *entities.User
		wantErr bool
	}{
		{
			name: "successful friendship creation",
			user1: &entities.User{
				ID:    1,
				Email: "andy@mail.com",
			},
			user2: &entities.User{
				ID:    2,
				Email: "alice@mail.com",
			},
			wantErr: false,
		},
		{
			name: "duplicate friendship should fail",
			user1: &entities.User{
				ID:    1,
				Email: "andy@mail.com",
			},
			user2: &entities.User{
				ID:    2,
				Email: "alice@mail.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateFriendship(tt.user1, tt.user2)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}

				// Verify friendship was created
				var count int
				query := "SELECT COUNT(*) FROM friends WHERE user1_id = $1 AND user2_id = $2"
				user1ID, user2ID := tt.user1.ID, tt.user2.ID
				if user1ID > user2ID {
					user1ID, user2ID = user2ID, user1ID
				}

				err = db.QueryRow(query, user1ID, user2ID).Scan(&count)
				if err != nil {
					t.Errorf("Failed to verify friendship: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected 1 friendship record, got %d", count)
				}
			}
		})
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		email   string
		wantErr bool
		wantID  int
	}{
		{
			name:    "existing user",
			email:   "andy@mail.com",
			wantErr: false,
			wantID:  1,
		},
		{
			name:    "non-existing user",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetUserByEmail(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if user == nil {
					t.Error("expected user, got nil")
				} else {
					if user.Email != tt.email {
						t.Errorf("expected email %s, got %s", tt.email, user.Email)
					}
					if user.ID != tt.wantID {
						t.Errorf("expected ID %d, got %d", tt.wantID, user.ID)
					}
				}
			}
		})
	}
}
