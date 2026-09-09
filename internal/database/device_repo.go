package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Device represents a row in the devices table.
type Device struct {
	ID             string    `json:"id"`
	Platform       string    `json:"platform"` // "android" or "ios"
	Model          string    `json:"model"`
	Manufacturer   string    `json:"manufacturer"`
	OSVersion      string    `json:"os_version"`
	APILevel       int       `json:"api_level,omitempty"`
	ScreenWidth    int       `json:"screen_width"`
	ScreenHeight   int       `json:"screen_height"`
	ScreenDensity  float64   `json:"screen_density,omitempty"`
	RAMMB          int       `json:"ram_mb,omitempty"`
	Features       string    `json:"features"`        // JSON array string
	FirstSeen      time.Time `json:"first_seen"`
	LastSeen       time.Time `json:"last_seen"`
	TotalSessions  int       `json:"total_sessions"`
	ConnectionType string    `json:"connection_type"` // "usb", "wifi", "emulator", "simulator"
	Nickname       string    `json:"nickname,omitempty"`
}

// DeviceRepository provides CRUD operations on the devices table.
type DeviceRepository struct {
	db *Database
}

// NewDeviceRepository creates a DeviceRepository.
func NewDeviceRepository(db *Database) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Upsert inserts a device or updates it if the ID already exists.
// On conflict, it preserves first_seen and total_sessions from the existing row.
func (r *DeviceRepository) Upsert(ctx context.Context, d Device) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO devices (id, platform, model, manufacturer, os_version, api_level,
			screen_width, screen_height, screen_density, ram_mb, features,
			connection_type, nickname, last_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			platform = excluded.platform,
			model = excluded.model,
			manufacturer = excluded.manufacturer,
			os_version = excluded.os_version,
			api_level = excluded.api_level,
			screen_width = excluded.screen_width,
			screen_height = excluded.screen_height,
			screen_density = excluded.screen_density,
			ram_mb = excluded.ram_mb,
			features = excluded.features,
			connection_type = excluded.connection_type,
			nickname = COALESCE(excluded.nickname, devices.nickname),
			last_seen = CURRENT_TIMESTAMP
	`, d.ID, d.Platform, d.Model, d.Manufacturer, d.OSVersion, d.APILevel,
		d.ScreenWidth, d.ScreenHeight, d.ScreenDensity, d.RAMMB, d.Features,
		d.ConnectionType, d.Nickname)
	if err != nil {
		return fmt.Errorf("upsert device %s: %w", d.ID, err)
	}
	return nil
}

// GetByID retrieves a single device by ID.
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*Device, error) {
	d := &Device{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, platform, COALESCE(model,''), COALESCE(manufacturer,''),
			COALESCE(os_version,''), COALESCE(api_level,0),
			COALESCE(screen_width,0), COALESCE(screen_height,0),
			COALESCE(screen_density,0), COALESCE(ram_mb,0),
			COALESCE(features,'[]'), first_seen, last_seen,
			COALESCE(total_sessions,0), COALESCE(connection_type,''),
			COALESCE(nickname,'')
		FROM devices WHERE id = ?
	`, id).Scan(
		&d.ID, &d.Platform, &d.Model, &d.Manufacturer,
		&d.OSVersion, &d.APILevel,
		&d.ScreenWidth, &d.ScreenHeight,
		&d.ScreenDensity, &d.RAMMB,
		&d.Features, &d.FirstSeen, &d.LastSeen,
		&d.TotalSessions, &d.ConnectionType,
		&d.Nickname,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get device %s: %w", id, err)
	}
	return d, nil
}

// List returns all known devices, ordered by last_seen descending.
func (r *DeviceRepository) List(ctx context.Context) ([]Device, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, platform, COALESCE(model,''), COALESCE(manufacturer,''),
			COALESCE(os_version,''), COALESCE(api_level,0),
			COALESCE(screen_width,0), COALESCE(screen_height,0),
			COALESCE(screen_density,0), COALESCE(ram_mb,0),
			COALESCE(features,'[]'), first_seen, last_seen,
			COALESCE(total_sessions,0), COALESCE(connection_type,''),
			COALESCE(nickname,'')
		FROM devices ORDER BY last_seen DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(
			&d.ID, &d.Platform, &d.Model, &d.Manufacturer,
			&d.OSVersion, &d.APILevel,
			&d.ScreenWidth, &d.ScreenHeight,
			&d.ScreenDensity, &d.RAMMB,
			&d.Features, &d.FirstSeen, &d.LastSeen,
			&d.TotalSessions, &d.ConnectionType,
			&d.Nickname,
		); err != nil {
			return nil, fmt.Errorf("scan device row: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// UpdateLastSeen bumps the last_seen timestamp for a device.
func (r *DeviceRepository) UpdateLastSeen(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE devices SET last_seen = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("update last_seen for %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device %s: %w", id, ErrNotFound)
	}
	return nil
}

// SetNickname updates the user-friendly nickname for a device.
func (r *DeviceRepository) SetNickname(ctx context.Context, id, nickname string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE devices SET nickname = ? WHERE id = ?", nickname, id)
	if err != nil {
		return fmt.Errorf("set nickname for %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device %s: %w", id, ErrNotFound)
	}
	return nil
}

// IncrementSessions atomically increments total_sessions and updates last_seen.
func (r *DeviceRepository) IncrementSessions(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE devices
		SET total_sessions = total_sessions + 1, last_seen = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("increment sessions for %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device %s: %w", id, ErrNotFound)
	}
	return nil
}

// Delete removes a device by ID. Cascades to device_quirks and sessions via FK.
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete device %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("device %s: %w", id, ErrNotFound)
	}
	return nil
}
