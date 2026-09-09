package database

import (
	"context"
	"errors"
	"testing"
)

func TestQuirkRepository_Lifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	devRepo := NewDeviceRepository(db)
	quirkRepo := NewQuirkRepository(db)
	ctx := context.Background()

	// Device must exist due to FK
	if err := devRepo.Upsert(ctx, Device{ID: "dev-q1", Platform: "android", ConnectionType: "usb"}); err != nil {
		t.Fatalf("upsert device: %v", err)
	}

	q := DeviceQuirk{
		DeviceID:    "dev-q1",
		QuirkType:   "slow_keyboard",
		Description: "Keyboard takes 500ms to open",
		Workaround:  "Add 500ms sleep before typing",
		Confidence:  0.8,
	}

	// 1. Upsert (Insert)
	if err := quirkRepo.Upsert(ctx, q); err != nil {
		t.Fatalf("upsert quirk: %v", err)
	}

	// 2. GetByDeviceAndType
	got, err := quirkRepo.GetByDeviceAndType(ctx, "dev-q1", "slow_keyboard")
	if err != nil {
		t.Fatalf("get quirk by type: %v", err)
	}
	if got.Description != q.Description {
		t.Errorf("expected description %q, got %q", q.Description, got.Description)
	}
	if got.Confidence != 0.8 {
		t.Errorf("expected confidence 0.8, got %f", got.Confidence)
	}

	// 3. GetByID
	byId, err := quirkRepo.GetByID(ctx, got.ID)
	if err != nil {
		t.Fatalf("get quirk by ID: %v", err)
	}
	if byId.QuirkType != "slow_keyboard" {
		t.Errorf("expected quirk type 'slow_keyboard', got %q", byId.QuirkType)
	}

	// 4. Upsert (Update on conflict)
	q.Description = "Updated description"
	q.Confidence = 0.95
	if err := quirkRepo.Upsert(ctx, q); err != nil {
		t.Fatalf("upsert quirk (update): %v", err)
	}

	updated, err := quirkRepo.GetByID(ctx, got.ID)
	if err != nil {
		t.Fatalf("get updated quirk: %v", err)
	}
	if updated.Description != "Updated description" {
		t.Errorf("expected updated description, got %q", updated.Description)
	}
	if updated.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", updated.Confidence)
	}

	// 5. UpdateConfidence
	if err := quirkRepo.UpdateConfidence(ctx, got.ID, 0.99); err != nil {
		t.Fatalf("update confidence: %v", err)
	}
	updated, _ = quirkRepo.GetByID(ctx, got.ID)
	if updated.Confidence != 0.99 {
		t.Errorf("expected confidence 0.99, got %f", updated.Confidence)
	}

	// 6. ListByDevice
	// Insert another quirk
	if err := quirkRepo.Upsert(ctx, DeviceQuirk{
		DeviceID:    "dev-q1",
		QuirkType:   "touch_jitter",
		Description: "Occasional tap miss",
		Confidence:  0.6,
	}); err != nil {
		t.Fatalf("upsert second quirk: %v", err)
	}

	list, err := quirkRepo.ListByDevice(ctx, "dev-q1")
	if err != nil {
		t.Fatalf("list quirks: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 quirks, got %d", len(list))
	}
	// Ordered by confidence DESC: first should be 0.99, second 0.6
	if list[0].Confidence < list[1].Confidence {
		t.Errorf("expected quirks ordered by confidence DESC, got %f then %f", list[0].Confidence, list[1].Confidence)
	}

	// 7. Delete
	if err := quirkRepo.Delete(ctx, got.ID); err != nil {
		t.Fatalf("delete quirk: %v", err)
	}
	_, err = quirkRepo.GetByID(ctx, got.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestQuirkRepository_ForeignKeyEnforcement(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	quirkRepo := NewQuirkRepository(db)
	ctx := context.Background()

	// Should fail because non-existent-device does not exist
	err := quirkRepo.Upsert(ctx, DeviceQuirk{
		DeviceID:    "non-existent-device",
		QuirkType:   "glitch",
		Description: "test",
		Confidence:  0.5,
	})
	if err == nil {
		t.Fatal("expected foreign key violation error, got nil")
	}
}

func TestQuirkRepository_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	quirkRepo := NewQuirkRepository(db)
	ctx := context.Background()

	_, err := quirkRepo.GetByID(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent ID, got %v", err)
	}

	_, err = quirkRepo.GetByDeviceAndType(ctx, "dev-none", "glitch")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent device/type, got %v", err)
	}

	err = quirkRepo.Delete(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on delete, got %v", err)
	}

	err = quirkRepo.UpdateConfidence(ctx, 99999, 0.5)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on update confidence, got %v", err)
	}
}
