package test

import (
	"bytes"
	"chat-service/api/proto"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/grpc"
)

// **Feature: gateway-grpc-implementation, Property 1: HTTP-to-gRPC Routing Correctness (Chat endpoints)**
// **Validates: Requirements 3.1-3.5**

// MockChatServiceServer implements the ChatService gRPC interface for testing
type MockChatServiceServer struct {
	proto.UnimplementedChatServiceServer
	LastMethod string
	LastRequest interface{}
}

func (m *MockChatServiceServer) GetAllChats(ctx context.Context, req *proto.GetAllChatsRequest) (*proto.GetAllChatsResponse, error) {
	m.LastMethod = "GetAllChats"
	m.LastRequest = req
	return &proto.GetAllChatsResponse{
		Chats: []*proto.Chat{{Id: 1, Name: "test"}},
		Total: 1,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

func (m *MockChatServiceServer) SearchChats(ctx context.Context, req *proto.SearchChatsRequest) (*proto.SearchChatsResponse, error) {
	m.LastMethod = "SearchChats"
	m.LastRequest = req
	return &proto.SearchChatsResponse{
		Chats: []*proto.Chat{{Id: 1, Name: "test"}},
		Total: 1,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

func (m *MockChatServiceServer) GetChatByID(ctx context.Context, req *proto.GetChatByIDRequest) (*proto.GetChatByIDResponse, error) {
	m.LastMethod = "GetChatByID"
	m.LastRequest = req
	return &proto.GetChatByIDResponse{
		Chat: &proto.Chat{Id: req.Id, Name: "test"},
	}, nil
}

func (m *MockChatServiceServer) CreateChat(ctx context.Context, req *proto.CreateChatRequest) (*proto.CreateChatResponse, error) {
	m.LastMethod = "CreateChat"
	m.LastRequest = req
	return &proto.CreateChatResponse{
		Chat: &proto.Chat{Id: 1, Name: req.Name},
	}, nil
}

func (m *MockChatServiceServer) GetAllAdministrators(ctx context.Context, req *proto.GetAllAdministratorsRequest) (*proto.GetAllAdministratorsResponse, error) {
	m.LastMethod = "GetAllAdministrators"
	m.LastRequest = req
	return &proto.GetAllAdministratorsResponse{
		Administrators: []*proto.Administrator{{Id: 1, ChatId: 1}},
		Total:          1,
		Page:           req.Page,
		Limit:          req.Limit,
	}, nil
}

func (m *MockChatServiceServer) GetAdministratorByID(ctx context.Context, req *proto.GetAdministratorByIDRequest) (*proto.GetAdministratorByIDResponse, error) {
	m.LastMethod = "GetAdministratorByID"
	m.LastRequest = req
	return &proto.GetAdministratorByIDResponse{
		Administrator: &proto.Administrator{Id: req.Id, ChatId: 1},
	}, nil
}

func (m *MockChatServiceServer) AddAdministrator(ctx context.Context, req *proto.AddAdministratorRequest) (*proto.AddAdministratorResponse, error) {
	m.LastMethod = "AddAdministrator"
	m.LastRequest = req
	return &proto.AddAdministratorResponse{
		Administrator: &proto.Administrator{Id: 1, ChatId: req.ChatId, Phone: req.Phone},
	}, nil
}

func (m *MockChatServiceServer) RemoveAdministrator(ctx context.Context, req *proto.RemoveAdministratorRequest) (*proto.RemoveAdministratorResponse, error) {
	m.LastMethod = "RemoveAdministrator"
	m.LastRequest = req
	return &proto.RemoveAdministratorResponse{Success: true}, nil
}

func (m *MockChatServiceServer) RefreshParticipantsCount(ctx context.Context, req *proto.RefreshParticipantsCountRequest) (*proto.RefreshParticipantsCountResponse, error) {
	m.LastMethod = "RefreshParticipantsCount"
	m.LastRequest = req
	return &proto.RefreshParticipantsCountResponse{ParticipantsCount: 10}, nil
}

func (m *MockChatServiceServer) AddAdministratorForMigration(ctx context.Context, req *proto.AddAdministratorForMigrationRequest) (*proto.AddAdministratorForMigrationResponse, error) {
	m.LastMethod = "AddAdministratorForMigration"
	m.LastRequest = req
	return &proto.AddAdministratorForMigrationResponse{
		Administrator: &proto.Administrator{Id: 1, ChatId: req.ChatId, Phone: req.Phone},
	}, nil
}

func (m *MockChatServiceServer) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	m.LastMethod = "Health"
	m.LastRequest = req
	return &proto.HealthResponse{Status: "OK"}, nil
}

// MockGatewayHandler simulates the Gateway Service HTTP handlers that route to gRPC
type MockGatewayHandler struct {
	chatClient proto.ChatServiceClient
}

func NewMockGatewayHandler(client proto.ChatServiceClient) *MockGatewayHandler {
	return &MockGatewayHandler{chatClient: client}
}

func (h *MockGatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Route based on HTTP method and path
	switch {
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/chats") && !strings.Contains(r.URL.Path, "/administrators"):
		if strings.Contains(r.URL.Path, "/search") {
			h.handleSearchChats(w, r, ctx)
		} else if len(strings.Split(strings.Trim(r.URL.Path, "/"), "/")) == 2 {
			h.handleGetChatByID(w, r, ctx)
		} else {
			h.handleGetAllChats(w, r, ctx)
		}
	case r.Method == "POST" && r.URL.Path == "/chats":
		h.handleCreateChat(w, r, ctx)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/administrators"):
		if len(strings.Split(strings.Trim(r.URL.Path, "/"), "/")) == 2 {
			h.handleGetAdministratorByID(w, r, ctx)
		} else {
			h.handleGetAllAdministrators(w, r, ctx)
		}
	case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/administrators"):
		h.handleAddAdministrator(w, r, ctx)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/administrators"):
		h.handleRemoveAdministrator(w, r, ctx)
	case r.Method == "POST" && strings.Contains(r.URL.Path, "/refresh-participants"):
		h.handleRefreshParticipantsCount(w, r, ctx)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *MockGatewayHandler) handleGetAllChats(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	req := &proto.GetAllChatsRequest{
		Page:      int32(page),
		Limit:     int32(limit),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}
	
	resp, err := h.chatClient.GetAllChats(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleSearchChats(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	req := &proto.SearchChatsRequest{
		Query:     r.URL.Query().Get("query"),
		Page:      int32(page),
		Limit:     int32(limit),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}
	
	resp, err := h.chatClient.SearchChats(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleGetChatByID(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id, _ := strconv.ParseInt(pathParts[1], 10, 64)
	
	req := &proto.GetChatByIDRequest{Id: id}
	resp, err := h.chatClient.GetChatByID(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleCreateChat(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req proto.CreateChatRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	resp, err := h.chatClient.CreateChat(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleGetAllAdministrators(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	req := &proto.GetAllAdministratorsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}
	
	resp, err := h.chatClient.GetAllAdministrators(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleGetAdministratorByID(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(pathParts[1], 10, 64)
	
	req := &proto.GetAdministratorByIDRequest{Id: id}
	resp, err := h.chatClient.GetAdministratorByID(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleAddAdministrator(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req proto.AddAdministratorRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	resp, err := h.chatClient.AddAdministrator(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleRemoveAdministrator(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id, _ := strconv.ParseInt(pathParts[1], 10, 64)
	
	req := &proto.RemoveAdministratorRequest{Id: id}
	resp, err := h.chatClient.RemoveAdministrator(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (h *MockGatewayHandler) handleRefreshParticipantsCount(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req proto.RefreshParticipantsCountRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	resp, err := h.chatClient.RefreshParticipantsCount(ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(resp)
}

// MockChatClient wraps the mock server to act as a gRPC client
type MockChatClient struct {
	server *MockChatServiceServer
}

func NewMockChatClient(server *MockChatServiceServer) *MockChatClient {
	return &MockChatClient{server: server}
}

func (c *MockChatClient) GetAllChats(ctx context.Context, req *proto.GetAllChatsRequest, opts ...grpc.CallOption) (*proto.GetAllChatsResponse, error) {
	return c.server.GetAllChats(ctx, req)
}

func (c *MockChatClient) SearchChats(ctx context.Context, req *proto.SearchChatsRequest, opts ...grpc.CallOption) (*proto.SearchChatsResponse, error) {
	return c.server.SearchChats(ctx, req)
}

func (c *MockChatClient) GetChatByID(ctx context.Context, req *proto.GetChatByIDRequest, opts ...grpc.CallOption) (*proto.GetChatByIDResponse, error) {
	return c.server.GetChatByID(ctx, req)
}

func (c *MockChatClient) CreateChat(ctx context.Context, req *proto.CreateChatRequest, opts ...grpc.CallOption) (*proto.CreateChatResponse, error) {
	return c.server.CreateChat(ctx, req)
}

func (c *MockChatClient) GetAllAdministrators(ctx context.Context, req *proto.GetAllAdministratorsRequest, opts ...grpc.CallOption) (*proto.GetAllAdministratorsResponse, error) {
	return c.server.GetAllAdministrators(ctx, req)
}

func (c *MockChatClient) GetAdministratorByID(ctx context.Context, req *proto.GetAdministratorByIDRequest, opts ...grpc.CallOption) (*proto.GetAdministratorByIDResponse, error) {
	return c.server.GetAdministratorByID(ctx, req)
}

func (c *MockChatClient) AddAdministrator(ctx context.Context, req *proto.AddAdministratorRequest, opts ...grpc.CallOption) (*proto.AddAdministratorResponse, error) {
	return c.server.AddAdministrator(ctx, req)
}

func (c *MockChatClient) RemoveAdministrator(ctx context.Context, req *proto.RemoveAdministratorRequest, opts ...grpc.CallOption) (*proto.RemoveAdministratorResponse, error) {
	return c.server.RemoveAdministrator(ctx, req)
}

func (c *MockChatClient) RefreshParticipantsCount(ctx context.Context, req *proto.RefreshParticipantsCountRequest, opts ...grpc.CallOption) (*proto.RefreshParticipantsCountResponse, error) {
	return c.server.RefreshParticipantsCount(ctx, req)
}

func (c *MockChatClient) AddAdministratorForMigration(ctx context.Context, req *proto.AddAdministratorForMigrationRequest, opts ...grpc.CallOption) (*proto.AddAdministratorForMigrationResponse, error) {
	return c.server.AddAdministratorForMigration(ctx, req)
}

func (c *MockChatClient) Health(ctx context.Context, req *proto.HealthRequest, opts ...grpc.CallOption) (*proto.HealthResponse, error) {
	return c.server.Health(ctx, req)
}

func TestChatGRPCRoutingCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: HTTP GET /chats routes to GetAllChats gRPC method
	properties.Property("HTTP GET /chats routes to GetAllChats gRPC method", prop.ForAll(
		func(page, limit int32, sortBy, sortOrder string) bool {
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request
			u := url.URL{Path: "/chats"}
			q := u.Query()
			q.Set("page", strconv.Itoa(int(page)))
			q.Set("limit", strconv.Itoa(int(limit)))
			q.Set("sort_by", sortBy)
			q.Set("sort_order", sortOrder)
			u.RawQuery = q.Encode()
			
			req := httptest.NewRequest("GET", u.String(), nil)
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "GetAllChats" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.GetAllChatsRequest)
			if !ok {
				return false
			}
			
			return grpcReq.Page == page && grpcReq.Limit == limit && 
				   grpcReq.SortBy == sortBy && grpcReq.SortOrder == sortOrder
		},
		gen.Int32Range(1, 100),
		gen.Int32Range(1, 50),
		gen.AlphaString(),
		gen.OneConstOf("asc", "desc", ""),
	))

	// Property: HTTP GET /chats/search routes to SearchChats gRPC method
	properties.Property("HTTP GET /chats/search routes to SearchChats gRPC method", prop.ForAll(
		func(query string, page, limit int32) bool {
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request
			u := url.URL{Path: "/chats/search"}
			q := u.Query()
			q.Set("query", query)
			q.Set("page", strconv.Itoa(int(page)))
			q.Set("limit", strconv.Itoa(int(limit)))
			u.RawQuery = q.Encode()
			
			req := httptest.NewRequest("GET", u.String(), nil)
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "SearchChats" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.SearchChatsRequest)
			if !ok {
				return false
			}
			
			return grpcReq.Query == query && grpcReq.Page == page && grpcReq.Limit == limit
		},
		gen.AlphaString(),
		gen.Int32Range(1, 100),
		gen.Int32Range(1, 50),
	))

	// Property: HTTP GET /chats/{id} routes to GetChatByID gRPC method
	properties.Property("HTTP GET /chats/{id} routes to GetChatByID gRPC method", prop.ForAll(
		func(chatID int64) bool {
			if chatID <= 0 {
				return true // Skip invalid IDs
			}
			
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request
			path := fmt.Sprintf("/chats/%d", chatID)
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "GetChatByID" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.GetChatByIDRequest)
			if !ok {
				return false
			}
			
			return grpcReq.Id == chatID
		},
		gen.Int64Range(1, 1000),
	))

	// Property: HTTP POST /chats routes to CreateChat gRPC method
	properties.Property("HTTP POST /chats routes to CreateChat gRPC method", prop.ForAll(
		func(name, url, maxChatID, source, department string, participantsCount int32) bool {
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request body
			reqBody := map[string]interface{}{
				"name":               name,
				"url":                url,
				"max_chat_id":        maxChatID,
				"source":             source,
				"department":         department,
				"participants_count": participantsCount,
			}
			
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/chats", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "CreateChat" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.CreateChatRequest)
			if !ok {
				return false
			}
			
			return grpcReq.Name == name && grpcReq.Url == url && 
				   grpcReq.MaxChatId == maxChatID && grpcReq.Source == source &&
				   grpcReq.Department == department && grpcReq.ParticipantsCount == participantsCount
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int32Range(0, 1000),
	))

	// Property: HTTP GET /administrators routes to GetAllAdministrators gRPC method
	properties.Property("HTTP GET /administrators routes to GetAllAdministrators gRPC method", prop.ForAll(
		func(page, limit int32) bool {
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request
			u := url.URL{Path: "/administrators"}
			q := u.Query()
			q.Set("page", strconv.Itoa(int(page)))
			q.Set("limit", strconv.Itoa(int(limit)))
			u.RawQuery = q.Encode()
			
			req := httptest.NewRequest("GET", u.String(), nil)
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "GetAllAdministrators" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.GetAllAdministratorsRequest)
			if !ok {
				return false
			}
			
			return grpcReq.Page == page && grpcReq.Limit == limit
		},
		gen.Int32Range(1, 100),
		gen.Int32Range(1, 50),
	))

	// Property: HTTP DELETE /administrators/{id} routes to RemoveAdministrator gRPC method
	properties.Property("HTTP DELETE /administrators/{id} routes to RemoveAdministrator gRPC method", prop.ForAll(
		func(adminID int64) bool {
			if adminID <= 0 {
				return true // Skip invalid IDs
			}
			
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request
			path := fmt.Sprintf("/administrators/%d", adminID)
			req := httptest.NewRequest("DELETE", path, nil)
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "RemoveAdministrator" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.RemoveAdministratorRequest)
			if !ok {
				return false
			}
			
			return grpcReq.Id == adminID
		},
		gen.Int64Range(1, 1000),
	))

	// Property: HTTP POST /refresh-participants routes to RefreshParticipantsCount gRPC method
	properties.Property("HTTP POST /refresh-participants routes to RefreshParticipantsCount gRPC method", prop.ForAll(
		func(chatID int64) bool {
			if chatID <= 0 {
				return true // Skip invalid IDs
			}
			
			mockServer := &MockChatServiceServer{}
			client := NewMockChatClient(mockServer)
			handler := NewMockGatewayHandler(client)
			
			// Create HTTP request body
			reqBody := map[string]interface{}{
				"chat_id": chatID,
			}
			
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/refresh-participants", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			// Execute HTTP request
			handler.ServeHTTP(w, req)
			
			// Verify correct gRPC method was called
			if mockServer.LastMethod != "RefreshParticipantsCount" {
				return false
			}
			
			// Verify request data was properly transformed
			grpcReq, ok := mockServer.LastRequest.(*proto.RefreshParticipantsCountRequest)
			if !ok {
				return false
			}
			
			return grpcReq.ChatId == chatID
		},
		gen.Int64Range(1, 1000),
	))

	// Run all properties with 100 iterations each
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}