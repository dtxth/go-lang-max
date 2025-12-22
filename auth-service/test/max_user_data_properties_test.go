package test

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/repository"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	_ "github.com/lib/pq"
)

// TestProperty4_UserDataManagementConsistency tests that MAX user data is properly managed
// **Feature: max-miniapp-auth, Property 4: User data management consistency**
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
func TestProperty4_UserDataManagementConsistency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(db.GetUnderlyingDB())

	userRepo := repository.NewUserPostgres(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test 1: Creating users with MAX fields stores all data correctly
	properties.Property("creating users with MAX fields stores all data correctly", prop.ForAll(
		func(phoneNum int, maxID int64, usernameSeed int, nameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			phone := "+7" + padNumber(phoneNum, 10)
			username := "user" + padNumber(usernameSeed, 6)
			name := "Name" + padNumber(nameSeed, 6)

			// Create user with MAX fields
			user := &domain.User{
				Phone:    phone,
				Email:    "",
				Password: "hashedpassword",
				Role:     domain.RoleOperator,
				MaxID:    &maxID,
				Username: &username,
				Name:     &name,
			}

			err := userRepo.Create(user)
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Retrieve user by max_id
			retrievedUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve user by max_id: %v", err)
				return false
			}

			// Verify all MAX fields are stored correctly
			if retrievedUser.MaxID == nil || *retrievedUser.MaxID != maxID {
				t.Logf("MaxID mismatch: expected %d, got %v", maxID, retrievedUser.MaxID)
				return false
			}

			if retrievedUser.Username == nil || *retrievedUser.Username != username {
				t.Logf("Username mismatch: expected %s, got %v", username, retrievedUser.Username)
				return false
			}

			if retrievedUser.Name == nil || *retrievedUser.Name != name {
				t.Logf("Name mismatch: expected %s, got %v", name, retrievedUser.Name)
				return false
			}

			// Verify other fields are also correct
			if retrievedUser.Phone != phone {
				t.Logf("Phone mismatch: expected %s, got %s", phone, retrievedUser.Phone)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number
		gen.Int64Range(1, 999999999),         // max_id
		gen.IntRange(100000, 999999),         // username seed
		gen.IntRange(100000, 999999),         // name seed
	))

	// Test 2: Creating users without username (optional field) handles gracefully
	properties.Property("creating users without username handles optional field gracefully", prop.ForAll(
		func(phoneNum int, maxID int64, nameSeed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			phone := "+7" + padNumber(phoneNum, 10)
			name := "Name" + padNumber(nameSeed, 6)

			// Create user with MAX fields but no username
			emptyUsername := ""
			user := &domain.User{
				Phone:    phone,
				Email:    "",
				Password: "hashedpassword",
				Role:     domain.RoleOperator,
				MaxID:    &maxID,
				Username: &emptyUsername, // Empty username
				Name:     &name,
			}

			err := userRepo.Create(user)
			if err != nil {
				t.Logf("Failed to create user without username: %v", err)
				return false
			}

			// Retrieve user by max_id
			retrievedUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve user by max_id: %v", err)
				return false
			}

			// Verify MAX fields are stored correctly
			if retrievedUser.MaxID == nil || *retrievedUser.MaxID != maxID {
				t.Logf("MaxID mismatch: expected %d, got %v", maxID, retrievedUser.MaxID)
				return false
			}

			if retrievedUser.Username == nil || *retrievedUser.Username != "" {
				t.Logf("Username should be empty, got %v", retrievedUser.Username)
				return false
			}

			if retrievedUser.Name == nil || *retrievedUser.Name != name {
				t.Logf("Name mismatch: expected %s, got %v", name, retrievedUser.Name)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number
		gen.Int64Range(1, 999999999),         // max_id
		gen.IntRange(100000, 999999),         // name seed
	))

	// Test 3: Updating users refreshes MAX fields correctly
	properties.Property("updating users refreshes MAX fields correctly", prop.ForAll(
		func(phoneNum int, maxID int64, username1Seed int, username2Seed int, name1Seed int, name2Seed int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			phone := "+7" + padNumber(phoneNum, 10)
			username1 := "user" + padNumber(username1Seed, 6)
			username2 := "user" + padNumber(username2Seed, 6)
			name1 := "Name" + padNumber(name1Seed, 6)
			name2 := "Name" + padNumber(name2Seed, 6)

			// Create user with initial MAX fields
			user := &domain.User{
				Phone:    phone,
				Email:    "",
				Password: "hashedpassword",
				Role:     domain.RoleOperator,
				MaxID:    &maxID,
				Username: &username1,
				Name:     &name1,
			}

			err := userRepo.Create(user)
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Update user with new MAX fields
			user.Username = &username2
			user.Name = &name2

			err = userRepo.Update(user)
			if err != nil {
				t.Logf("Failed to update user: %v", err)
				return false
			}

			// Retrieve user by max_id
			retrievedUser, err := userRepo.GetByMaxID(maxID)
			if err != nil {
				t.Logf("Failed to retrieve user by max_id: %v", err)
				return false
			}

			// Verify MAX fields are updated correctly
			if retrievedUser.Username == nil || *retrievedUser.Username != username2 {
				t.Logf("Username not updated: expected %s, got %v", username2, retrievedUser.Username)
				return false
			}

			if retrievedUser.Name == nil || *retrievedUser.Name != name2 {
				t.Logf("Name not updated: expected %s, got %v", name2, retrievedUser.Name)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number
		gen.Int64Range(1, 999999999),         // max_id
		gen.IntRange(100000, 999999),         // username1 seed
		gen.IntRange(100000, 999999),         // username2 seed
		gen.IntRange(100000, 999999),         // name1 seed
		gen.IntRange(100000, 999999),         // name2 seed
	))

	// Test 4: GetByMaxID returns correct user for lookup
	properties.Property("GetByMaxID returns correct user for lookup", prop.ForAll(
		func(phoneNum1 int, phoneNum2 int, maxID1 int64, maxID2 int64) bool {
			cleanupTestData(db.GetUnderlyingDB())

			// Ensure different phone numbers and max_ids
			if phoneNum1 == phoneNum2 || maxID1 == maxID2 {
				return true // Skip this test case
			}

			phone1 := "+7" + padNumber(phoneNum1, 10)
			phone2 := "+7" + padNumber(phoneNum2, 10)

			// Create two users with different MAX IDs
			username1 := "user1"
			name1 := "Name1"
			user1 := &domain.User{
				Phone:    phone1,
				Email:    "",
				Password: "hashedpassword1",
				Role:     domain.RoleOperator,
				MaxID:    &maxID1,
				Username: &username1,
				Name:     &name1,
			}

			username2 := "user2"
			name2 := "Name2"
			user2 := &domain.User{
				Phone:    phone2,
				Email:    "",
				Password: "hashedpassword2",
				Role:     domain.RoleOperator,
				MaxID:    &maxID2,
				Username: &username2,
				Name:     &name2,
			}

			err := userRepo.Create(user1)
			if err != nil {
				t.Logf("Failed to create user1: %v", err)
				return false
			}

			err = userRepo.Create(user2)
			if err != nil {
				t.Logf("Failed to create user2: %v", err)
				return false
			}

			// Retrieve user1 by max_id
			retrievedUser1, err := userRepo.GetByMaxID(maxID1)
			if err != nil {
				t.Logf("Failed to retrieve user1 by max_id: %v", err)
				return false
			}

			// Verify we got the correct user
			if retrievedUser1.Phone != phone1 {
				t.Logf("Retrieved wrong user: expected phone %s, got %s", phone1, retrievedUser1.Phone)
				return false
			}

			// Retrieve user2 by max_id
			retrievedUser2, err := userRepo.GetByMaxID(maxID2)
			if err != nil {
				t.Logf("Failed to retrieve user2 by max_id: %v", err)
				return false
			}

			// Verify we got the correct user
			if retrievedUser2.Phone != phone2 {
				t.Logf("Retrieved wrong user: expected phone %s, got %s", phone2, retrievedUser2.Phone)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number 1
		gen.IntRange(1000000000, 9999999999), // phone number 2
		gen.Int64Range(1, 999999999),         // max_id 1
		gen.Int64Range(1, 999999999),         // max_id 2
	))

	// Test 5: Users without MAX fields can still be created (backward compatibility)
	properties.Property("users without MAX fields can still be created", prop.ForAll(
		func(phoneNum int) bool {
			cleanupTestData(db.GetUnderlyingDB())

			phone := "+7" + padNumber(phoneNum, 10)

			// Create user without MAX fields
			emptyUsername := ""
			emptyName := ""
			user := &domain.User{
				Phone:    phone,
				Email:    "test@example.com",
				Password: "hashedpassword",
				Role:     domain.RoleOperator,
				MaxID:    nil, // No MAX ID
				Username: &emptyUsername,
				Name:     &emptyName,
			}

			err := userRepo.Create(user)
			if err != nil {
				t.Logf("Failed to create user without MAX fields: %v", err)
				return false
			}

			// Retrieve user by phone
			retrievedUser, err := userRepo.GetByPhone(phone)
			if err != nil {
				t.Logf("Failed to retrieve user by phone: %v", err)
				return false
			}

			// Verify MAX fields are nil/empty
			if retrievedUser.MaxID != nil {
				t.Logf("MaxID should be nil, got %v", retrievedUser.MaxID)
				return false
			}

			if retrievedUser.Username == nil || *retrievedUser.Username != "" {
				t.Logf("Username should be empty, got %v", retrievedUser.Username)
				return false
			}

			if retrievedUser.Name == nil || *retrievedUser.Name != "" {
				t.Logf("Name should be empty, got %v", retrievedUser.Name)
				return false
			}

			return true
		},
		gen.IntRange(1000000000, 9999999999), // phone number
	))

	properties.TestingRun(t)
}
