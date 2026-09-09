package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Session represents a row in the sessions table.
type Session struct {
	ID            string     `json:"id"`
	DeviceID      string     `json:"device_id"`
	AppPackage    string     `json:"app_package,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	EndedAt       *time.Time `json:"ended_at,omitempty"` // nil while session is active
	ActionsCount  int        `json:"actions_count"`
	FailuresCount int        `json:"failures_count"`
	Notes         string     `json:"notes,omitempty"`
}

// SessionRepository provides CRUD operations on the sessions table.
type SessionRepository struct {
	db *Database
}

// NewSessionRepository creates a SessionRepository.
func NewSessionRepository(db *Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create starts a new session.
func (r *SessionRepository) Create(ctx context.Context, s Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, device_id, app_package, notes)
		VALUES (?, ?, ?, ?)
	`, s.ID, s.DeviceID, s.AppPackage, s.Notes)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// End marks a session as ended, setting ended_at to now.
func (r *SessionRepository) End(ctx context.Context, sessionID string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET ended_at = CURRENT_TIMESTAMP WHERE id = ?
	`, sessionID)
	if err != nil {
		return fmt.Errorf("end session %s: %w", sessionID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session %s: %w", sessionID, ErrNotFound)
	}
	return nil
}

// GetByID retrieves a session by ID.
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*Session, error) {
	s := &Session{}
	var endedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, device_id, COALESCE(app_package,''), started_at, ended_at,
			actions_count, failures_count, COALESCE(notes,'')
		FROM sessions WHERE id = ?
	`, id).Scan(
		&s.ID, &s.DeviceID, &s.AppPackage, &s.StartedAt, &endedAt,
		&s.ActionsCount, &s.FailuresCount, &s.Notes,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get session %s: %w", id, err)
	}
	if endedAt.Valid {
		s.EndedAt = &endedAt.Time
	}
	return s, nil
}

// List returns sessions across all devices, most recent first.
// Pass limit=0 for default limit (100).
func (r *SessionRepository) List(ctx context.Context, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, device_id, COALESCE(app_package,''), started_at, ended_at,
			actions_count, failures_count, COALESCE(notes,'')
		FROM sessions
		ORDER BY started_at DESC
		LIMIT %d
	`, limit))
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// ListByDevice returns sessions for a device, most recent first.
// Pass limit=0 for no limit.
func (r *SessionRepository) ListByDevice(ctx context.Context, deviceID string, limit int) ([]Session, error) {
	query := `
		SELECT id, device_id, COALESCE(app_package,''), started_at, ended_at,
			actions_count, failures_count, COALESCE(notes,'')
		FROM sessions WHERE device_id = ?
		ORDER BY started_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list sessions for device %s: %w", deviceID, err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// scanRows scans multiple session rows.
func (r *SessionRepository) scanRows(rows *sql.Rows) ([]Session, error) {
	var sessions []Session
	for rows.Next() {
		var s Session
		var endedAt sql.NullTime
		if err := rows.Scan(
			&s.ID, &s.DeviceID, &s.AppPackage, &s.StartedAt, &endedAt,
			&s.ActionsCount, &s.FailuresCount, &s.Notes,
		); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		if endedAt.Valid {
			s.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// IncrementActions atomically increments the actions_count for a session.
func (r *SessionRepository) IncrementActions(ctx context.Context, sessionID string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE sessions SET actions_count = actions_count + 1 WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("increment actions for session %s: %w", sessionID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session %s: %w", sessionID, ErrNotFound)
	}
	return nil
}

// IncrementFailures atomically increments the failures_count for a session.
func (r *SessionRepository) IncrementFailures(ctx context.Context, sessionID string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE sessions SET failures_count = failures_count + 1 WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("increment failures for session %s: %w", sessionID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session %s: %w", sessionID, ErrNotFound)
	}
	return nil
}

// Delete removes a session by ID.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete session %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session %s: %w", id, ErrNotFound)
	}
	return nil
}
