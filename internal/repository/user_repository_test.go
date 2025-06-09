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


func TestUserRepository_GetFriendList(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Setup test data: Create some friendships first
	// The test data will be seeded by migrations, so we have users with IDs 1-5
	user1 := &entities.User{ID: 1, Email: "andy@mail.com"}
	user2 := &entities.User{ID: 2, Email: "alice@mail.com"}
	user3 := &entities.User{ID: 3, Email: "bob@mail.com"}
	user4 := &entities.User{ID: 4, Email: "jack@mail.com"}
	user5 := &entities.User{ID: 5, Email: "lisa@mail.com"}
	
	// Create an additional user for testing "no friends" case
	_, err := db.ExecContext(context.Background(), "INSERT INTO users (email) VALUES('nofriends@mail.com')")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	user6 := &entities.User{ID: 6, Email: "nofriends@mail.com"}

	// Create friendships for testing
	// andy (1) is friends with alice (2) and bob (3)
	err = repo.CreateFriendship(user1, user2)
	if err != nil {
		t.Fatalf("Failed to create friendship 1-2: %v", err)
	}
	err = repo.CreateFriendship(user1, user3)
	if err != nil {
		t.Fatalf("Failed to create friendship 1-3: %v", err)
	}

	// alice (2) is also friends with jack (4) - testing user2 as user1 in friendship table
	err = repo.CreateFriendship(user2, user4)
	if err != nil {
		t.Fatalf("Failed to create friendship 2-4: %v", err)
	}

	// bob (3) is friends with lisa (5) - testing user1 as user2 in friendship table
	err = repo.CreateFriendship(user5, user3) // This should store as (3,5) since 3 < 5
	if err != nil {
		t.Fatalf("Failed to create friendship 5-3: %v", err)
	}

	tests := []struct {
		name            string
		user            *entities.User
		expectedFriends []string // emails of expected friends, sorted
		wantErr         bool
	}{
		{
			name: "user with multiple friends",
			user: user1, // andy
			expectedFriends: []string{
				"alice@mail.com", // sorted alphabetically
				"bob@mail.com",
			},
			wantErr: false,
		},
		{
			name: "user with multiple friends (user as user1)",
			user: user2, // alice
			expectedFriends: []string{
				"andy@mail.com",
				"jack@mail.com",
			},
			wantErr: false,
		},
		{
			name: "user with multiple friends (user as user2)",
			user: user3, // bob
			expectedFriends: []string{
				"andy@mail.com",
				"lisa@mail.com",
			},
			wantErr: false,
		},
		{
			name: "user with one friend",
			user: user4, // jack
			expectedFriends: []string{
				"alice@mail.com",
			},
			wantErr: false,
		},
		{
			name: "user with one friend",
			user: user5, // lisa
			expectedFriends: []string{
				"bob@mail.com",
			},
			wantErr: false,
		},
		{
			name:            "user with no friends",
			user:            user6, // nofriends
			expectedFriends: []string{},
			wantErr:         false,
		},
		{
			name:    "non-existent user",
			user:    &entities.User{ID: 999, Email: "nonexistent@example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			friends, err := repo.GetFriendList(tt.user)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if len(friends) != len(tt.expectedFriends) {
				t.Errorf("expected %d friends, got %d", len(tt.expectedFriends), len(friends))
				return
			}

			// Check that friends are returned in alphabetical order by email
			for i, expectedEmail := range tt.expectedFriends {
				if friends[i].Email != expectedEmail {
					t.Errorf("expected friend %d to be %s, got %s", i, expectedEmail, friends[i].Email)
				}
			}

			// Verify no duplicate friends
			emailSet := make(map[string]bool)
			for _, friend := range friends {
				if emailSet[friend.Email] {
					t.Errorf("duplicate friend found: %s", friend.Email)
				}
				emailSet[friend.Email] = true
			}

			// Verify user is not in their own friend list
			for _, friend := range friends {
				if friend.ID == tt.user.ID {
					t.Errorf("user %s found in their own friend list", tt.user.Email)
				}
			}
		})
	}
}
