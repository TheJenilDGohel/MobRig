package database

import (
	"context"
	"errors"
	"testing"
)

func TestSessionRepository_Lifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Need a device first (FK constraint)
	devRepo := NewDeviceRepository(db)
	if err := devRepo.Upsert(ctx, Device{
		ID: "dev-1", Platform: "android", ConnectionType: "usb",
	}); err != nil {
		t.Fatalf("upsert device: %v", err)
	}

	repo := NewSessionRepository(db)

	// Create session
	if err := repo.Create(ctx, Session{
		ID:         "sess-001",
		DeviceID:   "dev-1",
		AppPackage: "com.example.app",
		Notes:      "Login flow test",
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Get it back
	s, err := repo.GetByID(ctx, "sess-001")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if s.AppPackage != "com.example.app" {
		t.Errorf("expected app_package 'com.example.app', got %q", s.AppPackage)
	}
	if s.EndedAt != nil {
		t.Error("expected ended_at to be nil for active session")
	}

	// Increment actions
	for range 3 {
		if err := repo.IncrementActions(ctx, "sess-001"); err != nil {
			t.Fatalf("increment actions: %v", err)
		}
	}

	// Increment failures
	if err := repo.IncrementFailures(ctx, "sess-001"); err != nil {
		t.Fatalf("increment failures: %v", err)
	}

	// End session
	if err := repo.End(ctx, "sess-001"); err != nil {
		t.Fatalf("end session: %v", err)
	}

	// Verify final state
	s, err = repo.GetByID(ctx, "sess-001")
	if err != nil {
		t.Fatalf("get ended session: %v", err)
	}
	if s.EndedAt == nil {
		t.Error("expected ended_at to be set after End()")
	}
	if s.ActionsCount != 3 {
		t.Errorf("expected 3 actions, got %d", s.ActionsCount)
	}
	if s.FailuresCount != 1 {
		t.Errorf("expected 1 failure, got %d", s.FailuresCount)
	}
}

func TestSessionRepository_ListByDevice(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()

	devRepo := NewDeviceRepository(db)
	if err := devRepo.Upsert(ctx, Device{
		ID: "dev-1", Platform: "android", ConnectionType: "usb",
	}); err != nil {
		t.Fatalf("upsert device: %v", err)
	}

	repo := NewSessionRepository(db)
	for _, id := range []string{"sess-a", "sess-b", "sess-c"} {
		if err := repo.Create(ctx, Session{
			ID:       id,
			DeviceID: "dev-1",
		}); err != nil {
			t.Fatalf("create session %s: %v", id, err)
		}
	}

	// List all
	sessions, err := repo.ListByDevice(ctx, "dev-1", 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}

	// List with limit
	sessions, err = repo.ListByDevice(ctx, "dev-1", 2)
	if err != nil {
		t.Fatalf("list with limit: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions with limit, got %d", len(sessions))
	}
}

func TestSessionRepository_ForeignKey(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db)

	// Creating a session with a non-existent device_id should fail (FK constraint)
	err := repo.Create(context.Background(), Session{
		ID:       "sess-orphan",
		DeviceID: "no-such-device",
	})
	if err == nil {
		t.Error("expected FK violation error for non-existent device")
	}
}

func TestSessionRepository_ListAllAndDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	devRepo := NewDeviceRepository(db)
	repo := NewSessionRepository(db)
	ctx := context.Background()

	// Insert 2 devices
	for _, id := range []string{"dev-a", "dev-b"} {
		if err := devRepo.Upsert(ctx, Device{ID: id, Platform: "android", ConnectionType: "usb"}); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	// Insert 1 session for dev-a, 2 for dev-b
	for _, s := range []Session{
		{ID: "sess-a1", DeviceID: "dev-a"},
		{ID: "sess-b1", DeviceID: "dev-b"},
		{ID: "sess-b2", DeviceID: "dev-b"},
	} {
		if err := repo.Create(ctx, s); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	// List across all devices
	all, err := repo.List(ctx, 10)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 total sessions, got %d", len(all))
	}

	// List with limit 2
	limited, err := repo.List(ctx, 2)
	if err != nil {
		t.Fatalf("list limited: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(limited))
	}

	// Delete
	if err := repo.Delete(ctx, "sess-a1"); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	_, err = repo.GetByID(ctx, "sess-a1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSessionRepository_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent session, got %v", err)
	}

	err = repo.End(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on End, got %v", err)
	}

	err = repo.IncrementActions(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on IncrementActions, got %v", err)
	}

	err = repo.IncrementFailures(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on IncrementFailures, got %v", err)
	}

	err = repo.Delete(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on Delete, got %v", err)
	}
}
