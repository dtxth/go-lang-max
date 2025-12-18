package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB with automatic reconnection
type DB struct {
	dsn        string
	db         *sql.DB
	mu         sync.RWMutex
	logger     *log.Logger
	maxRetries int
	retryDelay time.Duration
}

// NewDB creates a new database connection with automatic reconnection
func NewDB(dsn string, logger *log.Logger) *DB {
	if logger == nil {
		logger = log.New(log.Writer(), "[DB] ", log.LstdFlags)
	}
	
	return &DB{
		dsn:        dsn,
		logger:     logger,
		maxRetries: 3,
		retryDelay: time.Second * 2,
	}
}

// Connect establishes initial database connection
func (db *DB) Connect() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	return db.connectWithRetry()
}

// connectWithRetry attempts to connect with retry logic
func (db *DB) connectWithRetry() error {
	var err error
	
	for i := 0; i <= db.maxRetries; i++ {
		db.db, err = sql.Open("postgres", db.dsn)
		if err != nil {
			db.logger.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, db.maxRetries+1, err)
			if i < db.maxRetries {
				time.Sleep(db.retryDelay)
				continue
			}
			return fmt.Errorf("failed to open database after %d attempts: %w", db.maxRetries+1, err)
		}
		
		// Configure connection pool
		db.db.SetMaxOpenConns(25)
		db.db.SetMaxIdleConns(5)
		db.db.SetConnMaxLifetime(time.Hour) // Set reasonable lifetime
		
		// Test the connection
		if err = db.db.Ping(); err != nil {
			db.logger.Printf("Failed to ping database (attempt %d/%d): %v", i+1, db.maxRetries+1, err)
			db.db.Close()
			if i < db.maxRetries {
				time.Sleep(db.retryDelay)
				continue
			}
			return fmt.Errorf("failed to ping database after %d attempts: %w", db.maxRetries+1, err)
		}
		
		db.logger.Printf("Database connection established successfully (attempt %d/%d)", i+1, db.maxRetries+1)
		return nil
	}
	
	return err
}

// ensureConnection ensures database connection is alive, reconnecting if necessary
func (db *DB) ensureConnection() *sql.DB {
	db.mu.RLock()
	conn := db.db
	db.mu.RUnlock()
	
	// Check if connection is alive
	if conn != nil {
		if pingErr := conn.Ping(); pingErr == nil {
			return conn
		} else {
			db.logger.Printf("Database connection lost, attempting to reconnect: %v", pingErr)
		}
	}
	
	// Reconnect
	db.mu.Lock()
	defer db.mu.Unlock()
	
	// Double-check after acquiring write lock
	if db.db != nil {
		if err := db.db.Ping(); err == nil {
			return db.db
		}
		db.db.Close()
	}
	
	if err := db.connectWithRetry(); err != nil {
		db.logger.Printf("Failed to reconnect to database: %v", err)
		return nil
	}
	
	return db.db
}

// QueryRow executes a query that returns at most one row with automatic reconnection
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	conn := db.ensureConnection()
	if conn == nil {
		// Return a row that will produce an error when scanned
		return &sql.Row{}
	}
	return conn.QueryRow(query, args...)
}

// Query executes a query that returns rows with automatic reconnection
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	conn := db.ensureConnection()
	if conn == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	return conn.Query(query, args...)
}

// Exec executes a query without returning any rows with automatic reconnection
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	conn := db.ensureConnection()
	if conn == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	return conn.Exec(query, args...)
}

// Begin starts a transaction with automatic reconnection
func (db *DB) Begin() (*sql.Tx, error) {
	conn := db.ensureConnection()
	if conn == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	return conn.Begin()
}

// Close closes the database connection
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if db.db != nil {
		err := db.db.Close()
		db.db = nil
		return err
	}
	return nil
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	conn := db.ensureConnection()
	if conn == nil {
		return fmt.Errorf("no database connection available")
	}
	return conn.Ping()
}

// NewDBFromConnection wraps an existing sql.DB connection (useful for testing)
func NewDBFromConnection(sqlDB *sql.DB, logger *log.Logger) *DB {
	if logger == nil {
		logger = log.New(log.Writer(), "[DB] ", log.LstdFlags)
	}
	
	return &DB{
		db:         sqlDB,
		logger:     logger,
		maxRetries: 3,
		retryDelay: time.Second * 2,
	}
}

// GetUnderlyingDB returns the underlying *sql.DB (useful for testing and cleanup operations)
func (db *DB) GetUnderlyingDB() *sql.DB {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.db
}