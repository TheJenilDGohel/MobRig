package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// FailurePattern represents a row in the failure_patterns table.
type FailurePattern struct {
	ID          int64     `json:"id"`
	DeviceID    string    `json:"device_id,omitempty"` // nullable FK
	AppPackage  string    `json:"app_package,omitempty"`
	PatternType string    `json:"pattern_type"`
	Description string    `json:"description"`
	StackTrace  string    `json:"stack_trace,omitempty"`
	Frequency   int       `json:"frequency"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Resolved    bool      `json:"resolved"`
}

// FailurePatternRepository provides CRUD operations on the failure_patterns table.
type FailurePatternRepository struct {
	db *Database
}

// NewFailurePatternRepository creates a FailurePatternRepository.
func NewFailurePatternRepository(db *Database) *FailurePatternRepository {
	return &FailurePatternRepository{db: db}
}

// Create inserts a new failure pattern.
func (r *FailurePatternRepository) Create(ctx context.Context, fp FailurePattern) (int64, error) {
	// device_id is nullable, so pass sql.NullString.
	var deviceID sql.NullString
	if fp.DeviceID != "" {
		deviceID = sql.NullString{String: fp.DeviceID, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO failure_patterns (device_id, app_package, pattern_type, description, stack_trace)
		VALUES (?, ?, ?, ?, ?)
	`, deviceID, fp.AppPackage, fp.PatternType, fp.Description, fp.StackTrace)
	if err != nil {
		return 0, fmt.Errorf("create failure pattern: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

// ListByDevice returns failure patterns for a device, unresolved first.
func (r *FailurePatternRepository) ListByDevice(ctx context.Context, deviceID string) ([]FailurePattern, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(device_id,''), COALESCE(app_package,''),
			pattern_type, description, COALESCE(stack_trace,''),
			frequency, first_seen, last_seen, resolved
		FROM failure_patterns WHERE device_id = ?
		ORDER BY resolved ASC, last_seen DESC
	`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list failure patterns for device %s: %w", deviceID, err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListByApp returns failure patterns for an app package.
func (r *FailurePatternRepository) ListByApp(ctx context.Context, appPackage string) ([]FailurePattern, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(device_id,''), COALESCE(app_package,''),
			pattern_type, description, COALESCE(stack_trace,''),
			frequency, first_seen, last_seen, resolved
		FROM failure_patterns WHERE app_package = ?
		ORDER BY resolved ASC, last_seen DESC
	`, appPackage)
	if err != nil {
		return nil, fmt.Errorf("list failure patterns for app %s: %w", appPackage, err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// scanRows is a helper to scan multiple failure pattern rows.
func (r *FailurePatternRepository) scanRows(rows *sql.Rows) ([]FailurePattern, error) {
	var patterns []FailurePattern
	for rows.Next() {
		var fp FailurePattern
		if err := rows.Scan(
			&fp.ID, &fp.DeviceID, &fp.AppPackage,
			&fp.PatternType, &fp.Description, &fp.StackTrace,
			&fp.Frequency, &fp.FirstSeen, &fp.LastSeen, &fp.Resolved,
		); err != nil {
			return nil, fmt.Errorf("scan failure pattern row: %w", err)
		}
		patterns = append(patterns, fp)
	}
	return patterns, rows.Err()
}

// IncrementFrequency bumps the frequency count and updates last_seen.
func (r *FailurePatternRepository) IncrementFrequency(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE failure_patterns
		SET frequency = frequency + 1, last_seen = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("increment frequency for pattern %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("failure pattern %d: %w", id, ErrNotFound)
	}
	return nil
}

// Resolve marks a failure pattern as resolved.
func (r *FailurePatternRepository) Resolve(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE failure_patterns SET resolved = TRUE WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("resolve pattern %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("failure pattern %d: %w", id, ErrNotFound)
	}
	return nil
}

// GetByID retrieves a single failure pattern by ID.
func (r *FailurePatternRepository) GetByID(ctx context.Context, id int64) (*FailurePattern, error) {
	fp := &FailurePattern{}
	var deviceID sql.NullString
	var appPkg sql.NullString
	var stackTrace sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, device_id, app_package, pattern_type, description,
			stack_trace, frequency, first_seen, last_seen, resolved
		FROM failure_patterns WHERE id = ?
	`, id).Scan(
		&fp.ID, &deviceID, &appPkg, &fp.PatternType, &fp.Description,
		&stackTrace, &fp.Frequency, &fp.FirstSeen, &fp.LastSeen, &fp.Resolved,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("failure pattern %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get failure pattern %d: %w", id, err)
	}
	if deviceID.Valid {
		fp.DeviceID = deviceID.String
	}
	if appPkg.Valid {
		fp.AppPackage = appPkg.String
	}
	if stackTrace.Valid {
		fp.StackTrace = stackTrace.String
	}
	return fp, nil
}

// ListUnresolved returns all unresolved failure patterns across all devices.
// Pass limit=0 for default limit (50).
func (r *FailurePatternRepository) ListUnresolved(ctx context.Context, limit int) ([]FailurePattern, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, COALESCE(device_id,''), COALESCE(app_package,''),
			pattern_type, description, COALESCE(stack_trace,''),
			frequency, first_seen, last_seen, resolved
		FROM failure_patterns WHERE resolved = FALSE
		ORDER BY frequency DESC, last_seen DESC
		LIMIT %d
	`, limit))
	if err != nil {
		return nil, fmt.Errorf("list unresolved failure patterns: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// Delete removes a failure pattern by ID.
func (r *FailurePatternRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM failure_patterns WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete failure pattern %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("failure pattern %d: %w", id, ErrNotFound)
	}
	return nil
}
