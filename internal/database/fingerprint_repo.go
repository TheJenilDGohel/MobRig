package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ElementFingerprint represents a row in the element_fingerprints table.
type ElementFingerprint struct {
	ID             int64     `json:"id"`
	AppPackage     string    `json:"app_package"`
	ScreenName     string    `json:"screen_name,omitempty"`
	ElementText    string    `json:"element_text,omitempty"`
	ElementType    string    `json:"element_type,omitempty"`
	ResourceID     string    `json:"resource_id,omitempty"`
	ContentDesc    string    `json:"content_desc,omitempty"`
	StableSelector string    `json:"stable_selector"`
	LastSeen       time.Time `json:"last_seen"`
	Reliability    float64   `json:"reliability"`
}

// FingerprintRepository provides CRUD operations on the element_fingerprints table.
type FingerprintRepository struct {
	db *Database
}

// NewFingerprintRepository creates a FingerprintRepository.
func NewFingerprintRepository(db *Database) *FingerprintRepository {
	return &FingerprintRepository{db: db}
}

// Upsert inserts or updates an element fingerprint by (app_package, stable_selector).
func (r *FingerprintRepository) Upsert(ctx context.Context, f ElementFingerprint) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO element_fingerprints
			(app_package, screen_name, element_text, element_type,
			 resource_id, content_desc, stable_selector, reliability)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(app_package, stable_selector) DO UPDATE SET
			screen_name = excluded.screen_name,
			element_text = excluded.element_text,
			element_type = excluded.element_type,
			resource_id = excluded.resource_id,
			content_desc = excluded.content_desc,
			reliability = excluded.reliability,
			last_seen = CURRENT_TIMESTAMP
	`, f.AppPackage, f.ScreenName, f.ElementText, f.ElementType,
		f.ResourceID, f.ContentDesc, f.StableSelector, f.Reliability)
	if err != nil {
		return fmt.Errorf("upsert fingerprint for %s: %w", f.AppPackage, err)
	}
	return nil
}

// FindByApp returns all fingerprints for an app, ordered by reliability descending.
func (r *FingerprintRepository) FindByApp(ctx context.Context, appPackage string) ([]ElementFingerprint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_package, COALESCE(screen_name,''), COALESCE(element_text,''),
			COALESCE(element_type,''), COALESCE(resource_id,''),
			COALESCE(content_desc,''), stable_selector, last_seen, reliability
		FROM element_fingerprints WHERE app_package = ?
		ORDER BY reliability DESC
	`, appPackage)
	if err != nil {
		return nil, fmt.Errorf("find fingerprints for %s: %w", appPackage, err)
	}
	defer rows.Close()

	var fps []ElementFingerprint
	for rows.Next() {
		var f ElementFingerprint
		if err := rows.Scan(
			&f.ID, &f.AppPackage, &f.ScreenName, &f.ElementText,
			&f.ElementType, &f.ResourceID,
			&f.ContentDesc, &f.StableSelector, &f.LastSeen, &f.Reliability,
		); err != nil {
			return nil, fmt.Errorf("scan fingerprint row: %w", err)
		}
		fps = append(fps, f)
	}
	return fps, rows.Err()
}

// FindBySelector looks up a specific fingerprint by app and stable selector.
func (r *FingerprintRepository) FindBySelector(ctx context.Context, appPackage, selector string) (*ElementFingerprint, error) {
	f := &ElementFingerprint{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, app_package, COALESCE(screen_name,''), COALESCE(element_text,''),
			COALESCE(element_type,''), COALESCE(resource_id,''),
			COALESCE(content_desc,''), stable_selector, last_seen, reliability
		FROM element_fingerprints
		WHERE app_package = ? AND stable_selector = ?
	`, appPackage, selector).Scan(
		&f.ID, &f.AppPackage, &f.ScreenName, &f.ElementText,
		&f.ElementType, &f.ResourceID,
		&f.ContentDesc, &f.StableSelector, &f.LastSeen, &f.Reliability,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("fingerprint (%s, %s): %w", appPackage, selector, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find fingerprint: %w", err)
	}
	return f, nil
}

// GetByID retrieves a single fingerprint by its ID.
func (r *FingerprintRepository) GetByID(ctx context.Context, id int64) (*ElementFingerprint, error) {
	f := &ElementFingerprint{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, app_package, COALESCE(screen_name,''), COALESCE(element_text,''),
			COALESCE(element_type,''), COALESCE(resource_id,''),
			COALESCE(content_desc,''), stable_selector, last_seen, reliability
		FROM element_fingerprints
		WHERE id = ?
	`, id).Scan(
		&f.ID, &f.AppPackage, &f.ScreenName, &f.ElementText,
		&f.ElementType, &f.ResourceID,
		&f.ContentDesc, &f.StableSelector, &f.LastSeen, &f.Reliability,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("fingerprint %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get fingerprint: %w", err)
	}
	return f, nil
}

// UpdateReliability updates the reliability score for a fingerprint.
func (r *FingerprintRepository) UpdateReliability(ctx context.Context, id int64, reliability float64) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE element_fingerprints SET reliability = ?, last_seen = CURRENT_TIMESTAMP WHERE id = ?",
		reliability, id)
	if err != nil {
		return fmt.Errorf("update reliability for fingerprint %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("fingerprint %d: %w", id, ErrNotFound)
	}
	return nil
}

// Delete removes a fingerprint by ID.
func (r *FingerprintRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM element_fingerprints WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete fingerprint %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("fingerprint %d: %w", id, ErrNotFound)
	}
	return nil
}
