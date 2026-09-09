package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Flow represents a row in the flows table.
type Flow struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	AppPackage    string     `json:"app_package,omitempty"`
	DevicePlatform string   `json:"device_platform,omitempty"`
	Steps         string     `json:"steps"`          // JSON array of actions
	CreatedAt     time.Time  `json:"created_at"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	TotalRuns     int        `json:"total_runs"`
	SuccessCount  int        `json:"success_count"`
	AvgDurationMs int        `json:"avg_duration_ms"`
}

// FlowRepository provides CRUD operations on the flows table.
type FlowRepository struct {
	db *Database
}

// NewFlowRepository creates a FlowRepository.
func NewFlowRepository(db *Database) *FlowRepository {
	return &FlowRepository{db: db}
}

// Create inserts a new flow.
func (r *FlowRepository) Create(ctx context.Context, f Flow) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO flows (id, name, app_package, device_platform, steps)
		VALUES (?, ?, ?, ?, ?)
	`, f.ID, f.Name, f.AppPackage, f.DevicePlatform, f.Steps)
	if err != nil {
		return fmt.Errorf("create flow %q: %w", f.Name, err)
	}
	return nil
}

// GetByID retrieves a flow by ID.
func (r *FlowRepository) GetByID(ctx context.Context, id string) (*Flow, error) {
	return r.scanOne(ctx, "SELECT id, name, COALESCE(app_package,''), COALESCE(device_platform,''), steps, created_at, last_used_at, total_runs, success_count, COALESCE(avg_duration_ms,0) FROM flows WHERE id = ?", id)
}

// GetByName retrieves a flow by its unique name.
func (r *FlowRepository) GetByName(ctx context.Context, name string) (*Flow, error) {
	return r.scanOne(ctx, "SELECT id, name, COALESCE(app_package,''), COALESCE(device_platform,''), steps, created_at, last_used_at, total_runs, success_count, COALESCE(avg_duration_ms,0) FROM flows WHERE name = ?", name)
}

// scanOne is a helper to scan a single flow row.
func (r *FlowRepository) scanOne(ctx context.Context, query string, arg any) (*Flow, error) {
	f := &Flow{}
	var lastUsed sql.NullTime
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&f.ID, &f.Name, &f.AppPackage, &f.DevicePlatform, &f.Steps,
		&f.CreatedAt, &lastUsed, &f.TotalRuns, &f.SuccessCount, &f.AvgDurationMs,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flow: %w", ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get flow: %w", err)
	}
	if lastUsed.Valid {
		f.LastUsedAt = &lastUsed.Time
	}
	return f, nil
}

// List returns all flows, ordered by last_used_at descending (most recently used first).
func (r *FlowRepository) List(ctx context.Context) ([]Flow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(app_package,''), COALESCE(device_platform,''),
			steps, created_at, last_used_at, total_runs, success_count,
			COALESCE(avg_duration_ms,0)
		FROM flows ORDER BY COALESCE(last_used_at, created_at) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list flows: %w", err)
	}
	defer rows.Close()

	var flows []Flow
	for rows.Next() {
		var f Flow
		var lastUsed sql.NullTime
		if err := rows.Scan(
			&f.ID, &f.Name, &f.AppPackage, &f.DevicePlatform, &f.Steps,
			&f.CreatedAt, &lastUsed, &f.TotalRuns, &f.SuccessCount, &f.AvgDurationMs,
		); err != nil {
			return nil, fmt.Errorf("scan flow row: %w", err)
		}
		if lastUsed.Valid {
			f.LastUsedAt = &lastUsed.Time
		}
		flows = append(flows, f)
	}
	return flows, rows.Err()
}

// Update updates an existing flow's mutable fields (name, app_package, device_platform, steps).
func (r *FlowRepository) Update(ctx context.Context, f Flow) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE flows
		SET name = ?, app_package = ?, device_platform = ?, steps = ?
		WHERE id = ?
	`, f.Name, f.AppPackage, f.DevicePlatform, f.Steps, f.ID)
	if err != nil {
		return fmt.Errorf("update flow %s: %w", f.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("flow %s: %w", f.ID, ErrNotFound)
	}
	return nil
}

// RecordRun records a flow execution, updating run counts and rolling average duration.
func (r *FlowRepository) RecordRun(ctx context.Context, flowID string, success bool, durationMs int) error {
	// Use a rolling average: new_avg = ((old_avg * total_runs) + new_duration) / (total_runs + 1)
	successIncrement := 0
	if success {
		successIncrement = 1
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE flows SET
			total_runs = total_runs + 1,
			success_count = success_count + ?,
			avg_duration_ms = CASE
				WHEN total_runs = 0 THEN ?
				ELSE (avg_duration_ms * total_runs + ?) / (total_runs + 1)
			END,
			last_used_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, successIncrement, durationMs, durationMs, flowID)
	if err != nil {
		return fmt.Errorf("record run for flow %s: %w", flowID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("flow %s: %w", flowID, ErrNotFound)
	}
	return nil
}

// Delete removes a flow by ID.
func (r *FlowRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM flows WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete flow %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("flow %s: %w", id, ErrNotFound)
	}
	return nil
}
