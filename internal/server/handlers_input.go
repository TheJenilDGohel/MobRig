package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// ActionResponse is returned by mutation endpoints.
type ActionResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Duration string `json:"duration,omitempty"`
}

// TapRequest payload for POST /api/v1/devices/{id}/tap.
type TapRequest struct {
	X     *int `json:"x,omitempty"`
	Y     *int `json:"y,omitempty"`
	Index *int `json:"index,omitempty"`
}

func (s *Server) handleTap(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req TapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}

	x, y, err := s.resolveCoordinates(r.Context(), adapter, id, req.X, req.Y, req.Index)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	res, err := adapter.Tap(r.Context(), id, x, y)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Tap failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Tapped at (%d, %d)", x, y)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

// DoubleTapRequest payload for POST /api/v1/devices/{id}/double-tap.
type DoubleTapRequest struct {
	X     *int `json:"x,omitempty"`
	Y     *int `json:"y,omitempty"`
	Index *int `json:"index,omitempty"`
}

func (s *Server) handleDoubleTap(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req DoubleTapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}

	x, y, err := s.resolveCoordinates(r.Context(), adapter, id, req.X, req.Y, req.Index)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	res, err := adapter.DoubleTap(r.Context(), id, x, y)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Double tap failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Double-tapped at (%d, %d)", x, y)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

// LongPressRequest payload for POST /api/v1/devices/{id}/long-press.
type LongPressRequest struct {
	X          *int `json:"x,omitempty"`
	Y          *int `json:"y,omitempty"`
	Index      *int `json:"index,omitempty"`
	DurationMs int  `json:"duration_ms,omitempty"`
}

func (s *Server) handleLongPress(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req LongPressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}

	x, y, err := s.resolveCoordinates(r.Context(), adapter, id, req.X, req.Y, req.Index)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	durMs := 1000
	if req.DurationMs > 0 {
		durMs = req.DurationMs
	}

	res, err := adapter.LongPress(r.Context(), id, x, y, durMs)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Long press failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Long pressed at (%d, %d) for %dms", x, y, durMs)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

// SwipeRequest payload for POST /api/v1/devices/{id}/swipe.
type SwipeRequest struct {
	Direction  string   `json:"direction,omitempty"`
	Distance   *float64 `json:"distance,omitempty"`
	X1         *int     `json:"x1,omitempty"`
	Y1         *int     `json:"y1,omitempty"`
	X2         *int     `json:"x2,omitempty"`
	Y2         *int     `json:"y2,omitempty"`
	DurationMs int      `json:"duration_ms,omitempty"`
}

func (s *Server) handleSwipe(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req SwipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}

	durMs := 300
	if req.DurationMs > 0 {
		durMs = req.DurationMs
	}

	var x1, y1, x2, y2 int
	if req.Direction != "" {
		info, err := adapter.GetDeviceInfo(r.Context(), id)
		if err != nil {
			writeInternalError(w, fmt.Sprintf("Failed to get device screen size: %v", err))
			return
		}
		dist := 0.5
		if req.Distance != nil && *req.Distance > 0 && *req.Distance <= 1.0 {
			dist = *req.Distance
		}
		var calcErr error
		x1, y1, x2, y2, calcErr = s.calculateDirectionalSwipe(req.Direction, info.ScreenWidth, info.ScreenHeight, dist)
		if calcErr != nil {
			writeBadRequest(w, calcErr.Error())
			return
		}
	} else if req.X1 != nil && req.Y1 != nil && req.X2 != nil && req.Y2 != nil {
		x1, y1, x2, y2 = *req.X1, *req.Y1, *req.X2, *req.Y2
	} else {
		writeBadRequest(w, "Either direction or complete start/end coordinates (x1, y1, x2, y2) are required")
		return
	}

	res, err := adapter.Swipe(r.Context(), id, x1, y1, x2, y2, durMs)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Swipe failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Swiped from (%d, %d) to (%d, %d)", x1, y1, x2, y2)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

// TypeRequest payload for POST /api/v1/devices/{id}/type.
type TypeRequest struct {
	Text   string `json:"text"`
	Submit bool   `json:"submit,omitempty"`
}

func (s *Server) handleTypeText(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}

	res, err := adapter.TypeText(r.Context(), id, req.Text, req.Submit)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Type text failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Typed text %q", req.Text)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

// KeyRequest payload for POST /api/v1/devices/{id}/key.
type KeyRequest struct {
	KeyCode int `json:"key_code"`
}

func (s *Server) handlePressKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req KeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}

	if req.KeyCode <= 0 {
		writeBadRequest(w, "Valid key_code is required (e.g. 4 for Back, 66 for Enter)")
		return
	}

	res, err := adapter.PressButton(r.Context(), id, req.KeyCode)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Press button failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Pressed key code %d", req.KeyCode)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

func (s *Server) handleGoHome(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	res, err := adapter.GoHome(r.Context(), id)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Go home failed: %v", err))
		return
	}

	msg := "Navigated to home screen"
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

// --- Coordinate Helpers ---

func (s *Server) resolveCoordinates(ctx context.Context, adapter platform.PlatformAdapter, deviceID string, x, y, index *int) (int, int, error) {
	if x != nil && y != nil {
		return *x, *y, nil
	}
	if index != nil {
		tree, err := adapter.GetUITree(ctx, deviceID, platform.UITreeOptions{VisibleOnly: true})
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch UI tree for element index resolution: %w", err)
		}
		if tree != nil {
			for _, el := range tree.Elements {
				if el.Index == *index {
					return el.Bounds.CenterX(), el.Bounds.CenterY(), nil
				}
			}
		}
		return 0, 0, fmt.Errorf("element with index %d not found in current UI tree", *index)
	}
	return 0, 0, fmt.Errorf("either (x, y) coordinates or an element index are required")
}

func (s *Server) calculateDirectionalSwipe(direction string, screenW, screenH int, distance float64) (int, int, int, int, error) {
	if screenW <= 0 {
		screenW = 1080
	}
	if screenH <= 0 {
		screenH = 2400
	}

	midX := screenW / 2
	midY := screenH / 2

	switch strings.ToLower(direction) {
	case "up":
		delta := int(float64(screenH) * distance / 2)
		return midX, midY + delta, midX, midY - delta, nil
	case "down":
		delta := int(float64(screenH) * distance / 2)
		return midX, midY - delta, midX, midY + delta, nil
	case "left":
		delta := int(float64(screenW) * distance / 2)
		return midX + delta, midY, midX - delta, midY, nil
	case "right":
		delta := int(float64(screenW) * distance / 2)
		return midX - delta, midY, midX + delta, midY, nil
	default:
		return 0, 0, 0, 0, fmt.Errorf("unknown swipe direction %q (use 'up', 'down', 'left', or 'right')", direction)
	}
}
