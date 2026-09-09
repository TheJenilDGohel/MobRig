package database

import (
	"path/filepath"
	"testing"
)

func TestDatabaseOpenAndMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_mobrig.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Verify schema_meta table was populated by migration
	var version string
	err = db.QueryRow("SELECT value FROM schema_meta WHERE key = 'version'").Scan(&version)
	if err != nil {
		t.Fatalf("failed to query schema_meta: %v", err)
	}
	if version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %q", version)
	}

	// Verify devices table exists and allows insertion
	_, err = db.Exec(`
		INSERT INTO devices (id, platform, model, manufacturer, connection_type)
		VALUES ('test-device-1', 'android', 'Pixel 8', 'Google', 'usb')
	`)
	if err != nil {
		t.Fatalf("failed to insert test device: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count devices: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 device, got %d", count)
	}
}

func TestDatabaseWALMode(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Open(filepath.Join(tmpDir, "wal_test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected journal_mode 'wal', got %q", journalMode)
	}
}

func TestDatabaseForeignKeys(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := Open(filepath.Join(tmpDir, "fk_test.db"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		t.Fatalf("failed to query foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fkEnabled)
	}

	// Verify FK enforcement: inserting a session with invalid device_id should fail
	_, err = db.Exec(`
		INSERT INTO sessions (id, device_id, app_package)
		VALUES ('orphan-session', 'no-such-device', 'com.test.app')
	`)
	if err == nil {
		t.Error("expected FK violation error for session with invalid device_id")
	}
}

