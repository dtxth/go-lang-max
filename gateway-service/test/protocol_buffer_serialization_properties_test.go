package test

import (
	"testing"

	authpb "auth-service/api/proto"
	chatpb "chat-service/api/proto"
	employeepb "employee-service/api/proto"
	gatewaypb "gateway-service/api/proto"
	structurepb "structure-service/api/proto"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// TestProperty2_ProtocolBufferSerialization tests that Protocol Buffer serialization works correctly
// **Feature: gateway-grpc-implementation, Property 2: Protocol Buffer Serialization**
// **Validates: Requirements 1.2**
func TestProperty2_ProtocolBufferSerialization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test Auth Service messages
	properties.Property("auth service messages serialize/deserialize correctly", prop.ForAll(
		func(email, phone, password, role string, userID int64) bool {
			// Test RegisterRequest
			original := &authpb.RegisterRequest{
				Email:    email,
				Phone:    phone,
				Password: password,
				Role:     role,
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal RegisterRequest: %v", err)
				return false
			}

			// Deserialize
			deserialized := &authpb.RegisterRequest{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal RegisterRequest: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int64(),
	))

	// Test Chat Service messages
	properties.Property("chat service messages serialize/deserialize correctly", prop.ForAll(
		func(name, url, maxChatID, source, department string, participantsCount int32, chatID int64) bool {
			// Test CreateChatRequest
			original := &chatpb.CreateChatRequest{
				Name:              name,
				Url:               url,
				MaxChatId:         maxChatID,
				Source:            source,
				ParticipantsCount: participantsCount,
				Department:        department,
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal CreateChatRequest: %v", err)
				return false
			}

			// Deserialize
			deserialized := &chatpb.CreateChatRequest{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal CreateChatRequest: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int32(),
		gen.Int64(),
	))

	// Test Employee Service messages
	properties.Property("employee service messages serialize/deserialize correctly", prop.ForAll(
		func(firstName, lastName, middleName, phone, role string, universityID, employeeID int64) bool {
			// Test CreateEmployeeRequest
			original := &employeepb.CreateEmployeeRequest{
				FirstName:      firstName,
				LastName:       lastName,
				MiddleName:     middleName,
				Phone:          phone,
				Role:           role,
				UniversityName: "Test University",
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal CreateEmployeeRequest: %v", err)
				return false
			}

			// Deserialize
			deserialized := &employeepb.CreateEmployeeRequest{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal CreateEmployeeRequest: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int64(),
		gen.Int64(),
	))

	// Test Structure Service messages
	properties.Property("structure service messages serialize/deserialize correctly", prop.ForAll(
		func(name, inn, kpp, foiv string, universityID int64) bool {
			// Test CreateUniversityRequest
			original := &structurepb.CreateUniversityRequest{
				Name: name,
				Inn:  inn,
				Kpp:  kpp,
				Foiv: foiv,
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal CreateUniversityRequest: %v", err)
				return false
			}

			// Deserialize
			deserialized := &structurepb.CreateUniversityRequest{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal CreateUniversityRequest: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int64(),
	))

	// Test Gateway Service messages
	properties.Property("gateway service messages serialize/deserialize correctly", prop.ForAll(
		func(status string) bool {
			// Test HealthResponse
			original := &gatewaypb.HealthResponse{
				Status: status,
				Services: map[string]string{
					"auth":      "healthy",
					"chat":      "healthy",
					"employee":  "healthy",
					"structure": "healthy",
				},
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal HealthResponse: %v", err)
				return false
			}

			// Deserialize
			deserialized := &gatewaypb.HealthResponse{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal HealthResponse: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
	))

	// Test complex nested messages
	properties.Property("complex nested messages serialize/deserialize correctly", prop.ForAll(
		func(email, phone, accessToken, refreshToken string, userID int64) bool {
			// Test complex LoginResponse with nested TokenPair and User
			tokenPair := &authpb.TokenPair{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}

			user := &authpb.User{
				Id:        userID,
				Email:     email,
				Phone:     phone,
				Role:      "operator",
				CreatedAt: "2023-01-01T00:00:00Z",
			}

			original := &authpb.LoginResponse{
				Tokens: tokenPair,
				User:   user,
				Error:  "",
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal LoginResponse: %v", err)
				return false
			}

			// Deserialize
			deserialized := &authpb.LoginResponse{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal LoginResponse: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int64(),
	))

	// Test optional fields
	properties.Property("optional fields serialize/deserialize correctly", prop.ForAll(
		func(name, url, maxChatID, source, department string, participantsCount int32, hasUniversityID bool, universityID int64) bool {
			// Test CreateChatRequest with optional university_id
			original := &chatpb.CreateChatRequest{
				Name:              name,
				Url:               url,
				MaxChatId:         maxChatID,
				Source:            source,
				ParticipantsCount: participantsCount,
				Department:        department,
			}

			// Set optional field conditionally
			if hasUniversityID {
				original.UniversityId = &universityID
			}

			// Serialize
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal CreateChatRequest with optional field: %v", err)
				return false
			}

			// Deserialize
			deserialized := &chatpb.CreateChatRequest{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal CreateChatRequest with optional field: %v", err)
				return false
			}

			// Verify round-trip consistency
			return proto.Equal(original, deserialized)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int32(),
		gen.Bool(),
		gen.Int64(),
	))

	// Run all properties
	properties.TestingRun(t)
}