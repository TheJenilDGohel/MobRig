package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DeviceQuirk represents a row in the device_quirks table.
type DeviceQuirk struct {
	ID           int64     `json:"id"`
	DeviceID     string    `json:"device_id"`
	QuirkType    string    `json:"quirk_type"`
	Description  string    `json:"description"`
	Workaround   string    `json:"workaround,omitempty"`
	Confidence   float64   `json:"confidence"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

// QuirkRepository provides CRUD operations on the device_quirks table.
type QuirkRepository struct {
	db *Database
}

// NewQuirkRepository creates a QuirkRepository.
func NewQuirkRepository(db *Database) *QuirkRepository {
	return &QuirkRepository{db: db}
}

// Upsert inserts a quirk or updates it if (device_id, quirk_type) already exists.
func (r *QuirkRepository) Upsert(ctx context.Context, q DeviceQuirk) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO device_quirks (device_id, quirk_type, description, workaround, confidence)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(device_id, quirk_type) DO UPDATE SET
			description = excluded.description,
			workaround = excluded.workaround,
			confidence = excluded.confidence
	`, q.DeviceID, q.QuirkType, q.Description, q.Workaround, q.Confidence)
	if err != nil {
		return fmt.Errorf("upsert quirk for device %s: %w", q.DeviceID, err)
	}
	return nil
}

// ListByDevice returns all quirks for a given device, ordered by confidence descending.
func (r *QuirkRepository) ListByDevice(ctx context.Context, deviceID string) ([]DeviceQuirk, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, device_id, quirk_type, description,
			COALESCE(workaround,''), confidence, discovered_at
		FROM device_quirks WHERE device_id = ?
		ORDER BY confidence DESC
	`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list quirks for device %s: %w", deviceID, err)
	}
	defer rows.Close()

	var quirks []DeviceQuirk
	for rows.Next() {
		var q DeviceQuirk
		if err := rows.Scan(&q.ID, &q.DeviceID, &q.QuirkType, &q.Description,
			&q.Workaround, &q.Confidence, &q.DiscoveredAt); err != nil {
			return nil, fmt.Errorf("scan quirk row: %w", err)
		}
		quirks = append(quirks, q)
	}
	return quirks, rows.Err()
}

// UpdateConfidence sets the confidence score for a quirk.
func (r *QuirkRepository) UpdateConfidence(ctx context.Context, id int64, confidence float64) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE device_quirks SET confidence = ? WHERE id = ?", confidence, id)
	if err != nil {
		return fmt.Errorf("update quirk confidence: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("quirk %d: %w", id, ErrNotFound)
	}
	return nil
}

// Delete removes a quirk by ID.
func (r *QuirkRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM device_quirks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete quirk %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("quirk %d: %w", id, ErrNotFound)
	}
	return nil
}

// GetByID retrieves a single quirk by its ID.
func (r *QuirkRepository) GetByID(ctx context.Context, id int64) (*DeviceQuirk, error) {
	q := &DeviceQuirk{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, device_id, quirk_type, description,
			COALESCE(workaround,''), confidence, discovered_at
		FROM device_quirks WHERE id = ?
	`, id).Scan(
		&q.ID, &q.DeviceID, &q.QuirkType, &q.Description,
		&q.Workaround, &q.Confidence, &q.DiscoveredAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quirk %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get quirk: %w", err)
	}
	return q, nil
}

// GetByDeviceAndType retrieves a specific quirk by device ID and quirk type.
func (r *QuirkRepository) GetByDeviceAndType(ctx context.Context, deviceID, quirkType string) (*DeviceQuirk, error) {
	q := &DeviceQuirk{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, device_id, quirk_type, description,
			COALESCE(workaround,''), confidence, discovered_at
		FROM device_quirks WHERE device_id = ? AND quirk_type = ?
	`, deviceID, quirkType).Scan(
		&q.ID, &q.DeviceID, &q.QuirkType, &q.Description,
		&q.Workaround, &q.Confidence, &q.DiscoveredAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quirk (%s, %s): %w", deviceID, quirkType, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get quirk: %w", err)
	}
	return q, nil
}
