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

func TestUserRepository_GetCommonFriends(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Setup test data: Create users and friendships
	user1 := &entities.User{ID: 1, Email: "andy@mail.com"}    // andy
	user2 := &entities.User{ID: 2, Email: "alice@mail.com"}   // alice
	user3 := &entities.User{ID: 3, Email: "bob@mail.com"}     // bob
	user4 := &entities.User{ID: 4, Email: "jack@mail.com"}    // jack
	user5 := &entities.User{ID: 5, Email: "lisa@mail.com"}    // lisa

	// Create additional test users
	_, err := db.ExecContext(context.Background(), "INSERT INTO users (email) VALUES('charlie@mail.com'), ('diana@mail.com')")
	if err != nil {
		t.Fatalf("Failed to create test users: %v", err)
	}
	user6 := &entities.User{ID: 6, Email: "charlie@mail.com"}
	user7 := &entities.User{ID: 7, Email: "diana@mail.com"}

	// Create friendship network:
	// andy (1) <-> alice (2), bob (3), jack (4), charlie (6)
	// alice (2) <-> andy (1), bob (3), lisa (5), charlie (6)
	// bob (3) <-> andy (1), alice (2), charlie (6)
	// jack (4) <-> andy (1), alice (2)
	// lisa (5) <-> alice (2)
	// charlie (6) <-> andy (1), alice (2), bob (3)
	// diana (7) has no friends
	
	friendships := [][2]*entities.User{
		{user1, user2}, // andy <-> alice
		{user1, user3}, // andy <-> bob
		{user1, user4}, // andy <-> jack
		{user1, user6}, // andy <-> charlie
		{user2, user3}, // alice <-> bob
		{user2, user4}, // alice <-> jack
		{user2, user5}, // alice <-> lisa
		{user2, user6}, // alice <-> charlie
		{user3, user6}, // bob <-> charlie
	}

	for _, friendship := range friendships {
		err := repo.CreateFriendship(friendship[0], friendship[1])
		if err != nil {
			t.Fatalf("Failed to create friendship between %s and %s: %v", 
				friendship[0].Email, friendship[1].Email, err)
		}
	}

	tests := []struct {
		name            string
		user1           *entities.User
		user2           *entities.User
		expectedCommon  []string // emails of expected common friends, sorted
		wantErr         bool
	}{
		{
			name:  "users with multiple common friends",
			user1: user1, // andy: friends with alice, bob, jack, charlie
			user2: user2, // alice: friends with andy, bob, jack, lisa, charlie
			expectedCommon: []string{
				"bob@mail.com",     // bob is common friend of andy and alice
				"charlie@mail.com", // charlie is common friend of andy and alice
				"jack@mail.com",    // jack is common friend of andy and alice
			},
			wantErr: false,
		},
		{
			name:  "users with two common friends",
			user1: user1, // andy: friends with alice, bob, jack, charlie
			user2: user3, // bob: friends with andy, alice, charlie
			expectedCommon: []string{
				"alice@mail.com", // alice is common friend of andy and bob
				"charlie@mail.com", // charlie is common friend of andy and bob
			},
			wantErr: false,
		},
		{
			name:           "users with one common friend",
			user1:          user4, // jack: friends with andy, alice
			user2:          user5, // lisa: friends with alice
			expectedCommon: []string{"alice@mail.com"},
			wantErr:        false,
		},
		{
			name:           "one user has no friends",
			user1:          user1, // andy: has friends
			user2:          user7, // diana: no friends
			expectedCommon: []string{},
			wantErr:        false,
		},
		{
			name:           "user with no friends and user with friends",
			user1:          user7, // diana: no friends
			user2:          user6, // charlie: has friends but no common friends with diana
			expectedCommon: []string{},
			wantErr:        false,
		},
		{
			name:    "first user doesn't exist",
			user1:   &entities.User{ID: 999, Email: "nonexistent@example.com"},
			user2:   user1,
			wantErr: true,
		},
		{
			name:    "second user doesn't exist",
			user1:   user1,
			user2:   &entities.User{ID: 999, Email: "nonexistent@example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commonFriends, err := repo.GetCommonFriends(tt.user1, tt.user2)

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

			if len(commonFriends) != len(tt.expectedCommon) {
				t.Errorf("expected %d common friends, got %d", len(tt.expectedCommon), len(commonFriends))
				return
			}

			// Extract emails from returned friends and sort them for comparison
			var actualEmails []string
			for _, friend := range commonFriends {
				actualEmails = append(actualEmails, friend.Email)
			}
			sort.Strings(actualEmails)

			// Compare sorted expected vs actual
			expectedSorted := make([]string, len(tt.expectedCommon))
			copy(expectedSorted, tt.expectedCommon)
			sort.Strings(expectedSorted)

			for i, expectedEmail := range expectedSorted {
				if actualEmails[i] != expectedEmail {
					t.Errorf("expected common friend %d to be %s, got %s", i, expectedEmail, actualEmails[i])
				}
			}

			// Verify no duplicate common friends
			emailSet := make(map[string]bool)
			for _, friend := range commonFriends {
				if emailSet[friend.Email] {
					t.Errorf("duplicate common friend found: %s", friend.Email)
				}
				emailSet[friend.Email] = true
			}

			// Verify neither user is in their own common friends list
			for _, friend := range commonFriends {
				if friend.ID == tt.user1.ID || friend.ID == tt.user2.ID {
					t.Errorf("user found in common friends list: %s", friend.Email)
				}
			}
		})
	}
}
func TestUserRepository_CreateSubscription(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	tests := []struct {
		name       string
		subscriber *entities.User
		target     *entities.User
		wantErr    bool
	}{
		{
			name: "successful subscription creation",
			subscriber: &entities.User{
				ID:    1,
				Email: "andy@mail.com",
			},
			target: &entities.User{
				ID:    2,
				Email: "alice@mail.com",
			},
			wantErr: false,
		},
		{
			name: "duplicate subscription should fail",
			subscriber: &entities.User{
				ID:    1,
				Email: "andy@mail.com",
			},
			target: &entities.User{
				ID:    2,
				Email: "alice@mail.com",
			},
			wantErr: true,
		},
		{
			name: "subscription to self should fail",
			subscriber: &entities.User{
				ID:    1,
				Email: "andy@mail.com",
			},
			target: &entities.User{
				ID:    1,
				Email: "andy@mail.com",
			},
			wantErr: true,
		},
		{
			name: "subscription with different users",
			subscriber: &entities.User{
				ID:    3,
				Email: "bob@mail.com",
			},
			target: &entities.User{
				ID:    4,
				Email: "jack@mail.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateSubscription(tt.subscriber, tt.target)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}

				// Verify subscription was created
				var count int
				query := "SELECT COUNT(*) FROM subscriptions WHERE subscriber_id = $1 AND target_id = $2"
				err = db.QueryRow(query, tt.subscriber.ID, tt.target.ID).Scan(&count)
				if err != nil {
					t.Errorf("Failed to verify subscription: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected 1 subscription record, got %d", count)
				}
			}
		})
	}
}

func TestUserRepository_CreateBlockTx(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Setup test data - create some existing friendships and subscriptions
	user1 := &entities.User{ID: 1, Email: "andy@mail.com"}
	user2 := &entities.User{ID: 2, Email: "alice@mail.com"}
	user3 := &entities.User{ID: 3, Email: "bob@mail.com"}
	user4 := &entities.User{ID: 4, Email: "jack@mail.com"}

	// Create friendships between users 1-2 and 3-4
	err := repo.CreateFriendship(user1, user2)
	if err != nil {
		t.Fatalf("Failed to create friendship 1-2: %v", err)
	}
	err = repo.CreateFriendship(user3, user4)
	if err != nil {
		t.Fatalf("Failed to create friendship 3-4: %v", err)
	}

	// Create subscriptions
	err = repo.CreateSubscription(user1, user2) // user1 subscribes to user2
	if err != nil {
		t.Fatalf("Failed to create subscription 1->2: %v", err)
	}
	err = repo.CreateSubscription(user2, user1) // user2 subscribes to user1 (bidirectional)
	if err != nil {
		t.Fatalf("Failed to create subscription 2->1: %v", err)
	}
	err = repo.CreateSubscription(user3, user4) // user3 subscribes to user4
	if err != nil {
		t.Fatalf("Failed to create subscription 3->4: %v", err)
	}

	tests := []struct {
		name      string
		requestor *entities.User
		target    *entities.User
		wantErr   bool
		setup     func() // Additional setup for specific test cases
		verify    func(t *testing.T) // Verification logic for specific test cases
	}{
		{
			name:      "successful block creation with friendship and subscriptions cleanup",
			requestor: user1,
			target:    user2,
			wantErr:   false,
			verify: func(t *testing.T) {
				// Verify block was created
				var blockCount int
				err := db.QueryRow("SELECT COUNT(*) FROM blocks WHERE blocker_id = $1 AND blocked_id = $2", 
					user1.ID, user2.ID).Scan(&blockCount)
				if err != nil {
					t.Errorf("Failed to verify block: %v", err)
				}
				if blockCount != 1 {
					t.Errorf("Expected 1 block record, got %d", blockCount)
				}

				// Verify friendship was removed
				var friendshipCount int
				err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)", 
					user1.ID, user2.ID).Scan(&friendshipCount)
				if err != nil {
					t.Errorf("Failed to verify friendship removal: %v", err)
				}
				if friendshipCount != 0 {
					t.Errorf("Expected 0 friendship records, got %d", friendshipCount)
				}

				// Verify bidirectional subscriptions were removed
				var subscriptionCount int
				err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE (subscriber_id = $1 AND target_id = $2) OR (subscriber_id = $2 AND target_id = $1)", 
					user1.ID, user2.ID).Scan(&subscriptionCount)
				if err != nil {
					t.Errorf("Failed to verify subscription removal: %v", err)
				}
				if subscriptionCount != 0 {
					t.Errorf("Expected 0 subscription records, got %d", subscriptionCount)
				}
			},
		},
		{
			name:      "block creation without existing friendship",
			requestor: user3,
			target:    &entities.User{ID: 5, Email: "lisa@mail.com"}, // No existing friendship
			wantErr:   false,
			verify: func(t *testing.T) {
				// Verify block was created
				var blockCount int
				err := db.QueryRow("SELECT COUNT(*) FROM blocks WHERE blocker_id = $1 AND blocked_id = $2", 
					user3.ID, 5).Scan(&blockCount)
				if err != nil {
					t.Errorf("Failed to verify block: %v", err)
				}
				if blockCount != 1 {
					t.Errorf("Expected 1 block record, got %d", blockCount)
				}
			},
		},
		{
			name:      "duplicate block should fail",
			requestor: user1,
			target:    user2,
			wantErr:   true,
			setup: func() {
				// Block was already created in previous test
			},
		},
		{
			name:      "block creation with only one-way subscription",
			requestor: user4,
			target:    user3,
			wantErr:   false,
			verify: func(t *testing.T) {
				// Verify block was created
				var blockCount int
				err := db.QueryRow("SELECT COUNT(*) FROM blocks WHERE blocker_id = $1 AND blocked_id = $2", 
					user4.ID, user3.ID).Scan(&blockCount)
				if err != nil {
					t.Errorf("Failed to verify block: %v", err)
				}
				if blockCount != 1 {
					t.Errorf("Expected 1 block record, got %d", blockCount)
				}

				// Verify friendship was removed (user3-user4 friendship)
				var friendshipCount int
				err = db.QueryRow("SELECT COUNT(*) FROM friends WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)", 
					user3.ID, user4.ID).Scan(&friendshipCount)
				if err != nil {
					t.Errorf("Failed to verify friendship removal: %v", err)
				}
				if friendshipCount != 0 {
					t.Errorf("Expected 0 friendship records, got %d", friendshipCount)
				}

				// Verify subscription from user3 to user4 was removed
				var subscriptionCount int
				err = db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE subscriber_id = $1 AND target_id = $2", 
					user3.ID, user4.ID).Scan(&subscriptionCount)
				if err != nil {
					t.Errorf("Failed to verify subscription removal: %v", err)
				}
				if subscriptionCount != 0 {
					t.Errorf("Expected 0 subscription records from user3 to user4, got %d", subscriptionCount)
				}
			},
		},
		{
			name:      "block with invalid requestor",
			requestor: &entities.User{ID: 999, Email: "nonexistent@example.com"},
			target:    user2,
			wantErr:   true,
		},
		{
			name:      "block with invalid target",
			requestor: user2,
			target:    &entities.User{ID: 999, Email: "nonexistent@example.com"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := repo.CreateBlockTx(tt.requestor, tt.target)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}

				if tt.verify != nil {
					tt.verify(t)
				}
			}
		})
	}
}

func TestUserRepository_CheckBlockExists(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Setup test data - create a block
	user1 := &entities.User{ID: 1, Email: "andy@mail.com"}
	user2 := &entities.User{ID: 2, Email: "alice@mail.com"}
	user3 := &entities.User{ID: 3, Email: "bob@mail.com"}

	// Create a block from user1 to user2
	err := repo.CreateBlockTx(user1, user2)
	if err != nil {
		t.Fatalf("Failed to create block: %v", err)
	}

	tests := []struct {
		name        string
		requestorID int
		targetID    int
		expected    bool
		wantErr     bool
	}{
		{
			name:        "existing block",
			requestorID: user1.ID,
			targetID:    user2.ID,
			expected:    true,
			wantErr:     false,
		},
		{
			name:        "non-existing block (reverse)",
			requestorID: user2.ID,
			targetID:    user1.ID,
			expected:    false,
			wantErr:     false,
		},
		{
			name:        "non-existing block (different users)",
			requestorID: user1.ID,
			targetID:    user3.ID,
			expected:    false,
			wantErr:     false,
		},
		{
			name:        "non-existing block (completely different users)",
			requestorID: user2.ID,
			targetID:    user3.ID,
			expected:    false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.CheckBlockExists(tt.requestorID, tt.targetID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if exists != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, exists)
				}
			}
		})
	}
}

func TestUserRepository_CheckBidirectionalBlock(t *testing.T) {
	db, cleanup := setupTestContainer(t)
	defer cleanup()

	repo := NewUserRepository(db)

	// Setup test data - create blocks in different directions
	user1 := &entities.User{ID: 1, Email: "andy@mail.com"}
	user2 := &entities.User{ID: 2, Email: "alice@mail.com"}
	user3 := &entities.User{ID: 3, Email: "bob@mail.com"}
	user4 := &entities.User{ID: 4, Email: "jack@mail.com"}

	// Create block from user1 to user2
	err := repo.CreateBlockTx(user1, user2)
	if err != nil {
		t.Fatalf("Failed to create block 1->2: %v", err)
	}

	// Create block from user4 to user3
	err = repo.CreateBlockTx(user4, user3)
	if err != nil {
		t.Fatalf("Failed to create block 4->3: %v", err)
	}

	tests := []struct {
		name     string
		user1ID  int
		user2ID  int
		expected bool
		wantErr  bool
	}{
		{
			name:     "user1 blocks user2",
			user1ID:  user1.ID,
			user2ID:  user2.ID,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "user2 blocked by user1 (reverse check)",
			user1ID:  user2.ID,
			user2ID:  user1.ID,
			expected: true, // Should return true because user1 blocks user2
			wantErr:  false,
		},
		{
			name:     "user4 blocks user3",
			user1ID:  user4.ID,
			user2ID:  user3.ID,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "user3 blocked by user4 (reverse check)",
			user1ID:  user3.ID,
			user2ID:  user4.ID,
			expected: true, // Should return true because user4 blocks user3
			wantErr:  false,
		},
		{
			name:     "no block between user1 and user3",
			user1ID:  user1.ID,
			user2ID:  user3.ID,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "no block between user2 and user4",
			user1ID:  user2.ID,
			user2ID:  user4.ID,
			expected: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, err := repo.CheckBidirectionalBlock(tt.user1ID, tt.user2ID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if blocked != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, blocked)
				}
			}
		})
	}
}
