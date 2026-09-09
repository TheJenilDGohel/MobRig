package database

import (
	"context"
	"errors"
	"testing"
)

func TestFingerprintRepository_Lifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFingerprintRepository(db)
	ctx := context.Background()

	fp := ElementFingerprint{
		AppPackage:     "com.example.app",
		ScreenName:     "LoginScreen",
		ElementText:    "Sign In",
		ElementType:    "android.widget.Button",
		ResourceID:     "com.example.app:id/btn_login",
		ContentDesc:    "Login Button",
		StableSelector: "id:btn_login",
		Reliability:    0.9,
	}

	// 1. Upsert (Insert)
	if err := repo.Upsert(ctx, fp); err != nil {
		t.Fatalf("upsert fingerprint: %v", err)
	}

	// 2. FindBySelector
	got, err := repo.FindBySelector(ctx, "com.example.app", "id:btn_login")
	if err != nil {
		t.Fatalf("find by selector: %v", err)
	}
	if got.ElementText != "Sign In" {
		t.Errorf("expected ElementText 'Sign In', got %q", got.ElementText)
	}
	if got.Reliability != 0.9 {
		t.Errorf("expected Reliability 0.9, got %f", got.Reliability)
	}

	// 3. GetByID
	byId, err := repo.GetByID(ctx, got.ID)
	if err != nil {
		t.Fatalf("get by ID: %v", err)
	}
	if byId.StableSelector != "id:btn_login" {
		t.Errorf("expected selector 'id:btn_login', got %q", byId.StableSelector)
	}

	// 4. Upsert (Update on conflict: same app_package + stable_selector)
	fp.ElementText = "Log In"
	fp.Reliability = 0.99
	if err := repo.Upsert(ctx, fp); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	updated, err := repo.GetByID(ctx, got.ID)
	if err != nil {
		t.Fatalf("get updated: %v", err)
	}
	if updated.ElementText != "Log In" {
		t.Errorf("expected updated text 'Log In', got %q", updated.ElementText)
	}
	if updated.Reliability != 0.99 {
		t.Errorf("expected updated reliability 0.99, got %f", updated.Reliability)
	}

	// 5. UpdateReliability
	if err := repo.UpdateReliability(ctx, got.ID, 0.75); err != nil {
		t.Fatalf("update reliability: %v", err)
	}
	updated, _ = repo.GetByID(ctx, got.ID)
	if updated.Reliability != 0.75 {
		t.Errorf("expected reliability 0.75, got %f", updated.Reliability)
	}

	// 6. FindByApp
	// Add another fingerprint for same app
	if err := repo.Upsert(ctx, ElementFingerprint{
		AppPackage:     "com.example.app",
		ScreenName:     "LoginScreen",
		StableSelector: "id:btn_cancel",
		Reliability:    0.85,
	}); err != nil {
		t.Fatalf("upsert second fingerprint: %v", err)
	}

	appFps, err := repo.FindByApp(ctx, "com.example.app")
	if err != nil {
		t.Fatalf("find by app: %v", err)
	}
	if len(appFps) != 2 {
		t.Fatalf("expected 2 fingerprints, got %d", len(appFps))
	}
	// Ordered by reliability DESC: 0.85 should be before 0.75
	if appFps[0].Reliability < appFps[1].Reliability {
		t.Errorf("expected reliability DESC ordering, got %f then %f", appFps[0].Reliability, appFps[1].Reliability)
	}

	// 7. Delete
	if err := repo.Delete(ctx, got.ID); err != nil {
		t.Fatalf("delete fingerprint: %v", err)
	}
	_, err = repo.GetByID(ctx, got.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestFingerprintRepository_NotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	repo := NewFingerprintRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent ID, got %v", err)
	}

	_, err = repo.FindBySelector(ctx, "non.existent.app", "none")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent selector, got %v", err)
	}

	err = repo.Delete(ctx, 99999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on delete, got %v", err)
	}

	err = repo.UpdateReliability(ctx, 99999, 0.5)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on update reliability, got %v", err)
	}
}
