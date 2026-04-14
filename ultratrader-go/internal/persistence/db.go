package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents a connection to the SQLite database.
type DB struct {
	conn *sql.DB
}

// Connect initializes the database connection.
func Connect(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize schema
	if err := initSchema(conn); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// initSchema runs the initial DDL to setup the tables.
// We strictly use parameterized queries for all dynamic data,
// but DDL is static and therefore safe from injection.
func initSchema(conn *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS trade_exits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL,
		quantity REAL NOT NULL,
		exit_price REAL NOT NULL,
		profit_pct REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	return nil
}

// InsertTradeExit securely logs a completed trade into the database.
// It uses positional parameters (?) exclusively, unequivocally preventing SQL injection.
func (db *DB) InsertTradeExit(ctx context.Context, symbol, side string, quantity, exitPrice, profitPct float64) (int64, error) {
	// The query uses standard '?' parameter placeholders.
	// DO NOT use fmt.Sprintf to build SQL queries with dynamic data.
	query := `
		INSERT INTO trade_exits (symbol, side, quantity, exit_price, profit_pct)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.conn.ExecContext(ctx, query, symbol, side, quantity, exitPrice, profitPct)
	if err != nil {
		log.Printf("SQL Error: failed to insert trade exit for %s: %v", symbol, err)
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert id: %w", err)
	}

	return id, nil
}
