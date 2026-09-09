package database

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestDeviceRepository_UpsertAndGet(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	dev := Device{
		ID:             "emulator-5554",
		Platform:       "android",
		Model:          "Pixel 8",
		Manufacturer:   "Google",
		OSVersion:      "15",
		APILevel:       35,
		ScreenWidth:    1080,
		ScreenHeight:   2400,
		ScreenDensity:  2.75,
		RAMMB:          8192,
		Features:       `["nfc","bluetooth"]`,
		ConnectionType: "emulator",
	}

	// Insert
	if err := repo.Upsert(ctx, dev); err != nil {
		t.Fatalf("upsert device: %v", err)
	}

	// Get
	got, err := repo.GetByID(ctx, "emulator-5554")
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	if got.Model != "Pixel 8" {
		t.Errorf("expected model 'Pixel 8', got %q", got.Model)
	}
	if got.APILevel != 35 {
		t.Errorf("expected API level 35, got %d", got.APILevel)
	}
	if got.Platform != "android" {
		t.Errorf("expected platform 'android', got %q", got.Platform)
	}

	// Upsert again with updated model — should update, not duplicate
	dev.Model = "Pixel 9"
	if err := repo.Upsert(ctx, dev); err != nil {
		t.Fatalf("upsert device (update): %v", err)
	}
	got, err = repo.GetByID(ctx, "emulator-5554")
	if err != nil {
		t.Fatalf("get device after update: %v", err)
	}
	if got.Model != "Pixel 9" {
		t.Errorf("expected model 'Pixel 9' after update, got %q", got.Model)
	}
}

func TestDeviceRepository_List(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	// Insert two devices
	for _, d := range []Device{
		{ID: "dev-1", Platform: "android", Model: "Pixel 8", ConnectionType: "usb"},
		{ID: "dev-2", Platform: "android", Model: "Galaxy S24", ConnectionType: "wifi"},
	} {
		if err := repo.Upsert(ctx, d); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	devices, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
}

func TestDeviceRepository_UpdateLastSeen(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, Device{
		ID: "dev-1", Platform: "android", ConnectionType: "usb",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Should succeed
	if err := repo.UpdateLastSeen(ctx, "dev-1"); err != nil {
		t.Fatalf("update last seen: %v", err)
	}

	// Non-existent device should error
	if err := repo.UpdateLastSeen(ctx, "no-such-device"); err == nil {
		t.Error("expected error for non-existent device")
	}
}

func TestDeviceRepository_Delete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, Device{
		ID: "dev-del", Platform: "android", ConnectionType: "usb",
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := repo.Delete(ctx, "dev-del"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Get should fail
	if _, err := repo.GetByID(ctx, "dev-del"); err == nil {
		t.Error("expected error after delete")
	}
}

func TestDeviceRepository_SetNicknameAndIncrementSessions(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	if err := repo.Upsert(ctx, Device{ID: "dev-meta", Platform: "android", ConnectionType: "usb"}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Nickname
	if err := repo.SetNickname(ctx, "dev-meta", "My Test Pixel"); err != nil {
		t.Fatalf("set nickname: %v", err)
	}
	d, err := repo.GetByID(ctx, "dev-meta")
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	if d.Nickname != "My Test Pixel" {
		t.Errorf("expected nickname 'My Test Pixel', got %q", d.Nickname)
	}

	// Increment sessions
	if err := repo.IncrementSessions(ctx, "dev-meta"); err != nil {
		t.Fatalf("increment sessions: %v", err)
	}
	d, _ = repo.GetByID(ctx, "dev-meta")
	if d.TotalSessions != 1 {
		t.Errorf("expected total_sessions 1, got %d", d.TotalSessions)
	}
}

func TestDeviceRepository_GetNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for nonexistent device, got %v", err)
	}

	err = repo.SetNickname(ctx, "nonexistent", "nick")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for SetNickname on nonexistent device, got %v", err)
	}

	err = repo.IncrementSessions(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for IncrementSessions on nonexistent device, got %v", err)
	}

	err = repo.Delete(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for Delete on nonexistent device, got %v", err)
	}
}

// openTestDB is a test helper that creates a fresh database in a temp directory.
func openTestDB(t *testing.T) *Database {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	return db
}
