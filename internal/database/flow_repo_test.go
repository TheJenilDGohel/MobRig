package database

import (
	"context"
	"errors"
	"testing"
)

func TestFlowRepository_CRUD(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFlowRepository(db)
	ctx := context.Background()

	// Create
	if err := repo.Create(ctx, Flow{
		ID:             "flow-1",
		Name:           "login_flow",
		AppPackage:     "com.example.app",
		DevicePlatform: "android",
		Steps:          `[{"action":"tap","index":5},{"action":"type_text","text":"hello"}]`,
	}); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	// GetByID
	f, err := repo.GetByID(ctx, "flow-1")
	if err != nil {
		t.Fatalf("get flow by ID: %v", err)
	}
	if f.Name != "login_flow" {
		t.Errorf("expected name 'login_flow', got %q", f.Name)
	}

	// GetByName
	f, err = repo.GetByName(ctx, "login_flow")
	if err != nil {
		t.Fatalf("get flow by name: %v", err)
	}
	if f.ID != "flow-1" {
		t.Errorf("expected ID 'flow-1', got %q", f.ID)
	}

	// List
	flows, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list flows: %v", err)
	}
	if len(flows) != 1 {
		t.Errorf("expected 1 flow, got %d", len(flows))
	}

	// Delete
	if err := repo.Delete(ctx, "flow-1"); err != nil {
		t.Fatalf("delete flow: %v", err)
	}
	if _, err := repo.GetByID(ctx, "flow-1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestFlowRepository_Update(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFlowRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, Flow{
		ID:             "flow-u",
		Name:           "initial_flow",
		AppPackage:     "com.example.app",
		DevicePlatform: "android",
		Steps:          `[{"action":"tap"}]`,
	}); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	// Update mutable fields
	err := repo.Update(ctx, Flow{
		ID:             "flow-u",
		Name:           "renamed_flow",
		AppPackage:     "com.example.newapp",
		DevicePlatform: "ios",
		Steps:          `[{"action":"swipe"}]`,
	})
	if err != nil {
		t.Fatalf("update flow: %v", err)
	}

	updated, err := repo.GetByID(ctx, "flow-u")
	if err != nil {
		t.Fatalf("get updated flow: %v", err)
	}
	if updated.Name != "renamed_flow" {
		t.Errorf("expected name 'renamed_flow', got %q", updated.Name)
	}
	if updated.AppPackage != "com.example.newapp" {
		t.Errorf("expected package 'com.example.newapp', got %q", updated.AppPackage)
	}
	if updated.DevicePlatform != "ios" {
		t.Errorf("expected platform 'ios', got %q", updated.DevicePlatform)
	}
	if updated.Steps != `[{"action":"swipe"}]` {
		t.Errorf("expected steps updated, got %q", updated.Steps)
	}
}

func TestFlowRepository_RecordRun(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFlowRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, Flow{
		ID:    "flow-r",
		Name:  "record_test",
		Steps: `[]`,
	}); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	// Record 3 runs: 2 successful, 1 failed
	runs := []struct {
		success    bool
		durationMs int
	}{
		{true, 1000},
		{true, 2000},
		{false, 3000},
	}

	for _, run := range runs {
		if err := repo.RecordRun(ctx, "flow-r", run.success, run.durationMs); err != nil {
			t.Fatalf("record run: %v", err)
		}
	}

	f, err := repo.GetByID(ctx, "flow-r")
	if err != nil {
		t.Fatalf("get flow: %v", err)
	}
	if f.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", f.TotalRuns)
	}
	if f.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", f.SuccessCount)
	}
	if f.LastUsedAt == nil {
		t.Error("expected last_used_at to be set")
	}
	// Rolling average: (1000 + 2000 + 3000) / 3 = 2000
	// Due to integer arithmetic: (0*0 + 1000)/1 = 1000, (1000*1 + 2000)/2 = 1500, (1500*2 + 3000)/3 = 2000
	if f.AvgDurationMs != 2000 {
		t.Errorf("expected avg duration 2000, got %d", f.AvgDurationMs)
	}
}

func TestFlowRepository_UniqueName(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFlowRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, Flow{
		ID: "flow-1", Name: "unique_name", Steps: `[]`,
	}); err != nil {
		t.Fatalf("create flow: %v", err)
	}

	// Creating a second flow with the same name should fail
	err := repo.Create(ctx, Flow{
		ID: "flow-2", Name: "unique_name", Steps: `[]`,
	})
	if err == nil {
		t.Error("expected error for duplicate flow name")
	}
}

func TestFlowRepository_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFlowRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for GetByID on nonexistent, got %v", err)
	}

	_, err = repo.GetByName(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for GetByName on nonexistent, got %v", err)
	}

	err = repo.Update(ctx, Flow{ID: "nonexistent", Name: "foo"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for Update on nonexistent, got %v", err)
	}

	err = repo.RecordRun(ctx, "nonexistent", true, 100)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for RecordRun on nonexistent, got %v", err)
	}

	err = repo.Delete(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for Delete on nonexistent, got %v", err)
	}
}
