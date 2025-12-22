package http

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"gateway-service/internal/config"
	grpcClient "gateway-service/internal/infrastructure/grpc"
	"gateway-service/internal/infrastructure/middleware"
)

//go:embed docs
var docsFS embed.FS

// Router manages HTTP routing for the Gateway Service
type Router struct {
	handler *Handler
	mux     *http.ServeMux
	logger  *middleware.Logger
}

// NewRouter creates a new HTTP router
func NewRouter(cfg *config.Config, clientManager *grpcClient.ClientManager) *Router {
	handler := NewHandler(cfg, clientManager)
	mux := http.NewServeMux()

	// Initialize logger based on config
	logLevel := middleware.LogLevelInfo
	switch cfg.Logging.Level {
	case "debug":
		logLevel = middleware.LogLevelDebug
	case "warn":
		logLevel = middleware.LogLevelWarn
	case "error":
		logLevel = middleware.LogLevelError
	}
	logger := middleware.NewLogger(logLevel)

	router := &Router{
		handler: handler,
		mux:     mux,
		logger:  logger,
	}

	router.setupRoutes()
	return router
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Trace-ID")

	// Handle preflight requests
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Apply middleware chain and serve
	handler := http.Handler(r.mux)
	
	// Apply context propagation middleware
	handler = middleware.ContextPropagationMiddleware()(handler)
	
	// Apply request logging middleware
	handler = middleware.RequestLoggingMiddleware(r.logger)(handler)
	
	handler.ServeHTTP(w, req)
}

// setupRoutes configures all HTTP routes
func (r *Router) setupRoutes() {
	// Auth Service routes
	r.mux.HandleFunc("/register", r.handler.RegisterHandler)
	r.mux.HandleFunc("/login", r.handler.LoginHandler)
	r.mux.HandleFunc("/login-phone", r.handler.LoginByPhoneHandler)
	r.mux.HandleFunc("/refresh", r.handler.RefreshHandler)
	r.mux.HandleFunc("/logout", r.handler.LogoutHandler)
	r.mux.HandleFunc("/auth/max", r.handler.AuthenticateMAXHandler)
	r.mux.HandleFunc("/auth/password-reset/request", r.handler.RequestPasswordResetHandler)
	r.mux.HandleFunc("/auth/password-reset/confirm", r.handler.ResetPasswordHandler)
	r.mux.HandleFunc("/auth/password/change", r.handler.ChangePasswordHandler)
	r.mux.HandleFunc("/bot/me", r.handler.GetBotMeHandler)
	r.mux.HandleFunc("/metrics", r.handler.GetMetricsHandler)
	r.mux.HandleFunc("/health", r.handler.HealthHandler)

	// Chat Service routes
	r.mux.HandleFunc("/chats", r.chatRouteHandler)
	r.mux.HandleFunc("/chats/", r.chatRouteHandler)
	r.mux.HandleFunc("/chats/all", r.handler.GetAllChatsHandler)
	r.mux.HandleFunc("/chats/search", r.handler.SearchChatsHandler)
	r.mux.HandleFunc("/administrators", r.administratorRouteHandler)
	r.mux.HandleFunc("/administrators/", r.administratorRouteHandler)

	// Employee Service routes
	r.mux.HandleFunc("/employees/all", r.handler.GetAllEmployeesHandler)
	r.mux.HandleFunc("/employees/search", r.handler.SearchEmployeesHandler)
	r.mux.HandleFunc("/employees/", r.employeeRouteHandler)
	r.mux.HandleFunc("/employees/batch-update-maxid", r.handler.BatchUpdateMaxIDHandler)
	r.mux.HandleFunc("/employees/batch-status", r.handler.GetBatchStatusHandler)
	r.mux.HandleFunc("/employees/batch-status/", r.batchStatusRouteHandler)
	r.mux.HandleFunc("/simple-employee", r.handler.CreateEmployeeSimpleHandler)
	r.mux.HandleFunc("/create-employee", r.handler.CreateEmployeeHandler)

	// Structure Service routes
	r.mux.HandleFunc("/universities", r.universityRouteHandler)
	r.mux.HandleFunc("/universities/", r.universityRouteHandler)
	r.mux.HandleFunc("/structure", r.handler.CreateStructureHandler)
	r.mux.HandleFunc("/import/excel", r.handler.ImportExcelHandler)
	r.mux.HandleFunc("/branches/", r.branchRouteHandler)
	r.mux.HandleFunc("/faculties/", r.facultyRouteHandler)
	r.mux.HandleFunc("/groups/", r.groupRouteHandler)
	r.mux.HandleFunc("/departments/managers", r.departmentManagerRouteHandler)
	r.mux.HandleFunc("/departments/managers/", r.departmentManagerRouteHandler)

	// Swagger documentation routes
	r.mux.HandleFunc("/swagger/", r.swaggerRouteHandler)
	r.mux.HandleFunc("/swagger", r.swaggerRedirectHandler)
}

// chatRouteHandler handles chat-related routes with dynamic routing
func (r *Router) chatRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case path == "chats" && req.Method == http.MethodGet:
		r.handler.GetAllChatsHandler(w, req)
	case path == "chats" && req.Method == http.MethodPost:
		r.handler.CreateChatHandler(w, req)
	case len(parts) == 2 && parts[0] == "chats" && req.Method == http.MethodGet:
		r.handler.GetChatByIDHandler(w, req)
	case len(parts) == 3 && parts[0] == "chats" && parts[2] == "administrators" && req.Method == http.MethodPost:
		r.handler.AddAdministratorHandler(w, req)
	case len(parts) == 3 && parts[0] == "chats" && parts[2] == "refresh-participants" && req.Method == http.MethodPost:
		r.handler.RefreshParticipantsCountHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// administratorRouteHandler handles administrator-related routes
func (r *Router) administratorRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case path == "administrators" && req.Method == http.MethodGet:
		r.handler.GetAllAdministratorsHandler(w, req)
	case path == "administrators" && req.Method == http.MethodPost:
		r.handler.AddAdministratorHandler(w, req)
	case len(parts) == 2 && parts[0] == "administrators" && req.Method == http.MethodGet:
		r.handler.GetAdministratorByIDHandler(w, req)
	case len(parts) == 2 && parts[0] == "administrators" && req.Method == http.MethodDelete:
		r.handler.RemoveAdministratorHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// employeeRouteHandler handles employee-related routes with dynamic routing
func (r *Router) employeeRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 2 && parts[0] == "employees" && req.Method == http.MethodGet:
		r.handler.GetEmployeeByIDHandler(w, req)
	case len(parts) == 2 && parts[0] == "employees" && req.Method == http.MethodPut:
		r.handler.UpdateEmployeeHandler(w, req)
	case len(parts) == 2 && parts[0] == "employees" && req.Method == http.MethodDelete:
		r.handler.DeleteEmployeeHandler(w, req)
	case len(parts) == 2 && parts[0] == "employees" && req.Method == http.MethodPost:
		r.handler.CreateEmployeeHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// batchStatusRouteHandler handles batch status routes with dynamic routing
func (r *Router) batchStatusRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 3 && parts[0] == "employees" && parts[1] == "batch-status" && req.Method == http.MethodGet:
		r.handler.GetBatchStatusByIDHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// universityRouteHandler handles university-related routes with dynamic routing
func (r *Router) universityRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case path == "universities" && req.Method == http.MethodGet:
		r.handler.GetAllUniversitiesHandler(w, req)
	case path == "universities" && req.Method == http.MethodPost:
		r.handler.CreateUniversityHandler(w, req)
	case len(parts) == 2 && parts[0] == "universities" && req.Method == http.MethodGet:
		r.handler.GetUniversityByIDHandler(w, req)
	case len(parts) == 3 && parts[0] == "universities" && parts[2] == "structure" && req.Method == http.MethodGet:
		r.handler.GetUniversityStructureHandler(w, req)
	case len(parts) == 3 && parts[0] == "universities" && parts[2] == "name" && req.Method == http.MethodPut:
		r.handler.UpdateUniversityNameHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// departmentManagerRouteHandler handles department manager routes
func (r *Router) departmentManagerRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case path == "departments/managers" && req.Method == http.MethodGet:
		r.handler.GetAllDepartmentManagersHandler(w, req)
	case path == "departments/managers" && req.Method == http.MethodPost:
		r.handler.CreateDepartmentManagerHandler(w, req)
	case len(parts) == 3 && parts[0] == "departments" && parts[1] == "managers" && req.Method == http.MethodDelete:
		r.handler.RemoveDepartmentManagerHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// swaggerRedirectHandler redirects /swagger to /swagger/
func (r *Router) swaggerRedirectHandler(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
}

// swaggerRouteHandler serves swagger documentation files
func (r *Router) swaggerRouteHandler(w http.ResponseWriter, req *http.Request) {
	// Remove /swagger/ prefix
	path := strings.TrimPrefix(req.URL.Path, "/swagger/")
	
	// Default to index.html
	if path == "" {
		path = "index.html"
	}
	
	// Read file from embedded filesystem
	content, err := fs.ReadFile(docsFS, "docs/"+path)
	if err != nil {
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Documentation file not found", r.handler.getRequestID(req))
		return
	}
	
	// Set content type based on file extension
	contentType := "text/plain"
	if strings.HasSuffix(path, ".html") {
		contentType = "text/html; charset=utf-8"
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		contentType = "text/yaml; charset=utf-8"
	} else if strings.HasSuffix(path, ".json") {
		contentType = "application/json; charset=utf-8"
	} else if strings.HasSuffix(path, ".css") {
		contentType = "text/css; charset=utf-8"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript; charset=utf-8"
	}
	
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// branchRouteHandler handles branch-related routes
func (r *Router) branchRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 3 && parts[0] == "branches" && parts[2] == "name" && req.Method == http.MethodPut:
		r.handler.UpdateBranchNameHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// facultyRouteHandler handles faculty-related routes
func (r *Router) facultyRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 3 && parts[0] == "faculties" && parts[2] == "name" && req.Method == http.MethodPut:
		r.handler.UpdateFacultyNameHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}

// groupRouteHandler handles group-related routes
func (r *Router) groupRouteHandler(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 3 && parts[0] == "groups" && parts[2] == "name" && req.Method == http.MethodPut:
		r.handler.UpdateGroupNameHandler(w, req)
	case len(parts) == 3 && parts[0] == "groups" && parts[2] == "chat" && req.Method == http.MethodPut:
		r.handler.LinkGroupToChatHandler(w, req)
	default:
		r.handler.writeErrorResponse(w, http.StatusNotFound, "not_found", "Endpoint not found", r.handler.getRequestID(req))
	}
}