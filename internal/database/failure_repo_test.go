package database

import (
	"context"
	"errors"
	"testing"
)

func TestFailurePatternRepository_Lifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	devRepo := NewDeviceRepository(db)
	repo := NewFailurePatternRepository(db)
	ctx := context.Background()

	// Device for FK
	if err := devRepo.Upsert(ctx, Device{ID: "dev-f1", Platform: "android", ConnectionType: "usb"}); err != nil {
		t.Fatalf("upsert device: %v", err)
	}

	fp := FailurePattern{
		DeviceID:    "dev-f1",
		AppPackage:  "com.example.app",
		PatternType: "anr",
		Description: "Application Not Responding during image load",
		StackTrace:  "java.lang.RuntimeException: timeout",
	}

	// 1. Create
	id, err := repo.Create(ctx, fp)
	if err != nil {
		t.Fatalf("create failure pattern: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	// 2. GetByID
	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get by ID: %v", err)
	}
	if got.PatternType != "anr" {
		t.Errorf("expected pattern_type 'anr', got %q", got.PatternType)
	}
	if got.Frequency != 1 {
		t.Errorf("expected initial frequency 1, got %d", got.Frequency)
	}
	if got.Resolved {
		t.Error("expected resolved to be false initially")
	}

	// 3. IncrementFrequency
	if err := repo.IncrementFrequency(ctx, id); err != nil {
		t.Fatalf("increment frequency: %v", err)
	}
	got, _ = repo.GetByID(ctx, id)
	if got.Frequency != 2 {
		t.Errorf("expected frequency 2 after increment, got %d", got.Frequency)
	}

	// 4. Resolve
	if err := repo.Resolve(ctx, id); err != nil {
		t.Fatalf("resolve pattern: %v", err)
	}
	got, _ = repo.GetByID(ctx, id)
	if !got.Resolved {
		t.Error("expected resolved to be true after Resolve()")
	}

	// 5. ListByDevice
	devPatterns, err := repo.ListByDevice(ctx, "dev-f1")
	if err != nil {
		t.Fatalf("list by device: %v", err)
	}
	if len(devPatterns) != 1 {
		t.Errorf("expected 1 pattern for device, got %d", len(devPatterns))
	}

	// 6. ListByApp
	appPatterns, err := repo.ListByApp(ctx, "com.example.app")
	if err != nil {
		t.Fatalf("list by app: %v", err)
	}
	if len(appPatterns) != 1 {
		t.Errorf("expected 1 pattern for app, got %d", len(appPatterns))
	}

	// 7. ListUnresolved
	// Create another unresolved pattern
	id2, err := repo.Create(ctx, FailurePattern{
		AppPackage:  "com.example.app",
		PatternType: "crash",
		Description: "NullPointerException",
	})
	if err != nil {
		t.Fatalf("create second pattern: %v", err)
	}

	unresolved, err := repo.ListUnresolved(ctx, 10)
	if err != nil {
		t.Fatalf("list unresolved: %v", err)
	}
	if len(unresolved) != 1 {
		t.Fatalf("expected 1 unresolved pattern, got %d", len(unresolved))
	}
	if unresolved[0].ID != id2 {
		t.Errorf("expected unresolved ID %d, got %d", id2, unresolved[0].ID)
	}

	// 8. Delete
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("delete pattern: %v", err)
	}
	_, err = repo.GetByID(ctx, id)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestFailurePatternRepository_DeviceSetNullOnDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	devRepo := NewDeviceRepository(db)
	repo := NewFailurePatternRepository(db)
	ctx := context.Background()

	// Insert device
	if err := devRepo.Upsert(ctx, Device{ID: "dev-cascade", Platform: "android", ConnectionType: "usb"}); err != nil {
		t.Fatalf("upsert device: %v", err)
	}

	// Insert failure pattern referencing device
	id, err := repo.Create(ctx, FailurePattern{
		DeviceID:    "dev-cascade",
		PatternType: "oom",
		Description: "Out of memory",
	})
	if err != nil {
		t.Fatalf("create pattern: %v", err)
	}

	// Delete device — FK has ON DELETE SET NULL
	if err := devRepo.Delete(ctx, "dev-cascade"); err != nil {
		t.Fatalf("delete device: %v", err)
	}

	// Failure pattern should still exist, with device_id now empty (NULL in DB)
	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get pattern after device delete: %v", err)
	}
	if got.DeviceID != "" {
		t.Errorf("expected empty device_id after device deletion, got %q", got.DeviceID)
	}
}

func TestFailurePatternRepository_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFailurePatternRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent ID, got %v", err)
	}

	err = repo.IncrementFrequency(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on increment, got %v", err)
	}

	err = repo.Resolve(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on resolve, got %v", err)
	}

	err = repo.Delete(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on delete, got %v", err)
	}
}
