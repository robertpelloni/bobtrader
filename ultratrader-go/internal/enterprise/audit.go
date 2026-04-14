package enterprise

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditEvent represents a single logged action within the system.
type AuditEvent struct {
	Timestamp int64                  `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details"`
	PrevHash  string                 `json:"prev_hash"`
	Hash      string                 `json:"hash"`
}

// AuditLogger appends events to a file with cryptographic checksums chaining them together.
type AuditLogger struct {
	mu       sync.Mutex
	file     *os.File
	lastHash string
}

// NewAuditLogger initializes an append-only JSONL audit log.
func NewAuditLogger(filepath string) (*AuditLogger, error) {
	// Open file for appending, creating if it doesn't exist
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	// For a fresh start or simple implementation, initialize with an empty hash.
	// In production, you would read the last line of the file to recover the lastHash.
	return &AuditLogger{
		file:     f,
		lastHash: "0000000000000000000000000000000000000000000000000000000000000000",
	}, nil
}

// Close gracefully closes the audit log file.
func (a *AuditLogger) Close() error {
	if a.file != nil {
		return a.file.Close()
	}
	return nil
}

// Log securely appends an action to the audit trail, computing a SHA-256 chain hash.
func (a *AuditLogger) Log(userID, action, resource string, details map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	event := AuditEvent{
		Timestamp: time.Now().UnixNano(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Details:   details,
		PrevHash:  a.lastHash,
	}

	// Create JSON without the hash field for calculation
	rawEvent, err := json.Marshal(struct {
		Timestamp int64  `json:"timestamp"`
		UserID    string `json:"user_id"`
		Action    string `json:"action"`
		Resource  string `json:"resource"`
		PrevHash  string `json:"prev_hash"`
	}{
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
		Action:    event.Action,
		Resource:  event.Resource,
		PrevHash:  event.PrevHash,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal event for hashing: %w", err)
	}

	// Append stringified details to hash payload deterministically
	detailsStr := fmt.Sprintf("%v", details)
	hashPayload := append(rawEvent, []byte(detailsStr)...)

	// Compute SHA-256
	h := sha256.New()
	h.Write(hashPayload)
	event.Hash = hex.EncodeToString(h.Sum(nil))

	// Update last hash
	a.lastHash = event.Hash

	// Serialize final event with hash
	finalJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal final event: %w", err)
	}

	// Append to file
	if _, err := a.file.Write(append(finalJSON, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}
