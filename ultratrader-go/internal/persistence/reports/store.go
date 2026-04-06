package reports

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Report struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("report path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create report store directory: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) Append(_ context.Context, report Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if report.Timestamp.IsZero() {
		report.Timestamp = time.Now().UTC()
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open report store: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(report); err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	return nil
}

func (s *Store) Latest(limit int) ([]Report, error) {
	if limit <= 0 {
		return nil, nil
	}
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open report store for read: %w", err)
	}
	defer f.Close()

	var all []Report
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var report Report
		if err := json.Unmarshal(scanner.Bytes(), &report); err != nil {
			continue
		}
		all = append(all, report)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan report store: %w", err)
	}
	if len(all) <= limit {
		return all, nil
	}
	return all[len(all)-limit:], nil
}

func (s *Store) LatestByType() (map[string]Report, error) {
	reports, err := s.Latest(1000)
	if err != nil {
		return nil, err
	}
	out := make(map[string]Report)
	for _, report := range reports {
		out[report.Type] = report
	}
	return out, nil
}

func (s *Store) Path() string { return s.path }
