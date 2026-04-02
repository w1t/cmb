package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/codematicbench/cmb/pkg/agent"
)

// Store provides persistent storage for task results using SQLite
type Store struct {
	db *sql.DB
}

// NewStore creates a new SQLite storage instance
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// initSchema creates the database tables
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent TEXT NOT NULL,
		task TEXT NOT NULL,
		config TEXT,
		success BOOLEAN NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		duration_seconds REAL NOT NULL,
		error TEXT,
		output TEXT,
		evaluation JSON,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_runs_agent ON runs(agent);
	CREATE INDEX IF NOT EXISTS idx_runs_task ON runs(task);
	CREATE INDEX IF NOT EXISTS idx_runs_success ON runs(success);
	CREATE INDEX IF NOT EXISTS idx_runs_created_at ON runs(created_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveResult saves a task execution result to the database
func (s *Store) SaveResult(result *agent.Result) error {
	// Serialize evaluation as JSON
	var evalJSON []byte
	var err error
	if result.Evaluation != nil {
		evalJSON, err = json.Marshal(result.Evaluation)
		if err != nil {
			return fmt.Errorf("failed to marshal evaluation: %w", err)
		}
	}

	query := `
	INSERT INTO runs (agent, task, config, success, start_time, end_time, duration_seconds, error, output, evaluation)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		result.Agent,
		result.Task,
		result.ConfigUsed,
		result.Success,
		result.StartTime,
		result.EndTime,
		result.Duration.Seconds(),
		result.Error,
		result.Output,
		evalJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert result: %w", err)
	}

	return nil
}

// GetResults retrieves results with optional filters
func (s *Store) GetResults(filters map[string]interface{}, limit int) ([]*agent.Result, error) {
	query := "SELECT id, agent, task, config, success, start_time, end_time, duration_seconds, error, output, evaluation, created_at FROM runs WHERE 1=1"
	args := []interface{}{}

	// Add filters
	if agent, ok := filters["agent"].(string); ok && agent != "" {
		query += " AND agent = ?"
		args = append(args, agent)
	}
	if task, ok := filters["task"].(string); ok && task != "" {
		query += " AND task = ?"
		args = append(args, task)
	}
	if success, ok := filters["success"].(bool); ok {
		query += " AND success = ?"
		args = append(args, success)
	}

	// Order by most recent first
	query += " ORDER BY created_at DESC"

	// Add limit
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	results := []*agent.Result{}
	for rows.Next() {
		var r agent.Result
		var id int64
		var createdAt time.Time
		var durationSeconds float64
		var evalJSON []byte

		err := rows.Scan(
			&id,
			&r.Agent,
			&r.Task,
			&r.ConfigUsed,
			&r.Success,
			&r.StartTime,
			&r.EndTime,
			&durationSeconds,
			&r.Error,
			&r.Output,
			&evalJSON,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		r.Duration = time.Duration(durationSeconds * float64(time.Second))

		// Unmarshal evaluation if present
		if len(evalJSON) > 0 {
			var eval agent.EvalResult
			if err := json.Unmarshal(evalJSON, &eval); err != nil {
				return nil, fmt.Errorf("failed to unmarshal evaluation: %w", err)
			}
			r.Evaluation = &eval
		}

		results = append(results, &r)
	}

	return results, rows.Err()
}

// GetTaskStats returns statistics for a specific task
func (s *Store) GetTaskStats(taskName string) (map[string]interface{}, error) {
	query := `
	SELECT
		COUNT(*) as total_runs,
		SUM(CASE WHEN success THEN 1 ELSE 0 END) as successful_runs,
		AVG(duration_seconds) as avg_duration,
		MIN(duration_seconds) as min_duration,
		MAX(duration_seconds) as max_duration
	FROM runs
	WHERE task = ?
	`

	var stats struct {
		TotalRuns      int
		SuccessfulRuns int
		AvgDuration    float64
		MinDuration    float64
		MaxDuration    float64
	}

	err := s.db.QueryRow(query, taskName).Scan(
		&stats.TotalRuns,
		&stats.SuccessfulRuns,
		&stats.AvgDuration,
		&stats.MinDuration,
		&stats.MaxDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	result := map[string]interface{}{
		"total_runs":      stats.TotalRuns,
		"successful_runs": stats.SuccessfulRuns,
		"success_rate":    float64(stats.SuccessfulRuns) / float64(stats.TotalRuns),
		"avg_duration":    stats.AvgDuration,
		"min_duration":    stats.MinDuration,
		"max_duration":    stats.MaxDuration,
	}

	return result, nil
}
