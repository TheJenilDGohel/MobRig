package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// DeviceItemResponse describes a device in the list response.
type DeviceItemResponse struct {
	ID             string `json:"id"`
	Platform       string `json:"platform"`
	Model          string `json:"model"`
	Status         string `json:"status"`
	ConnectionType string `json:"connection_type"`
}

// DevicesListResponse is returned by GET /api/v1/devices.
type DevicesListResponse struct {
	Devices []DeviceItemResponse `json:"devices"`
	Count   int                  `json:"count"`
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.registry.ListAllDevices(r.Context())
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Failed to list devices: %v", err))
		return
	}

	items := make([]DeviceItemResponse, 0, len(devices))
	for _, d := range devices {
		items = append(items, DeviceItemResponse{
			ID:             d.ID,
			Platform:       string(d.Platform),
			Model:          d.Model,
			Status:         string(d.Status),
			ConnectionType: string(d.ConnectionType),
		})
	}

	writeJSON(w, http.StatusOK, DevicesListResponse{
		Devices: items,
		Count:   len(items),
	})
}

// DeviceDetailResponse is returned by GET /api/v1/devices/{id}.
type DeviceDetailResponse struct {
	Device *platform.DeviceInfo `json:"device"`
}

func (s *Server) handleGetDevice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	info, err := adapter.GetDeviceInfo(r.Context(), id)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Failed to get device info: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, DeviceDetailResponse{Device: info})
}

// ScreenshotResponse is returned by GET /api/v1/devices/{id}/screenshot when format=json.
type ScreenshotResponse struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SizeBytes int    `json:"size_bytes"`
	Format    string `json:"format"`
	Data      string `json:"data"` // base64 encoded PNG
}

func (s *Server) handleGetScreenshot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	shot, err := adapter.GetScreenshot(r.Context(), id)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Failed to capture screenshot: %v", err))
		return
	}

	// If raw binary requested via Accept header or query param
	accept := r.Header.Get("Accept")
	rawQuery := r.URL.Query().Get("raw")
	if strings.Contains(accept, "image/png") || rawQuery == "true" || rawQuery == "1" {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(shot.Data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(shot.Data)
		return
	}

	dataStr := shot.Base64
	if dataStr == "" && len(shot.Data) > 0 {
		dataStr = base64.StdEncoding.EncodeToString(shot.Data)
	}

	writeJSON(w, http.StatusOK, ScreenshotResponse{
		Width:     shot.Width,
		Height:    shot.Height,
		SizeBytes: len(shot.Data),
		Format:    "png",
		Data:      dataStr,
	})
}

// UITreeResponse is returned by GET /api/v1/devices/{id}/ui-tree.
type UITreeResponse struct {
	Tree       *platform.UITree `json:"tree"`
	TokenCount int              `json:"token_count"`
}

func (s *Server) handleGetUITree(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	visibleOnly := true
	if vo := r.URL.Query().Get("visible_only"); vo == "false" || vo == "0" {
		visibleOnly = false
	}
	textFilter := r.URL.Query().Get("text_filter")

	tree, err := adapter.GetUITree(r.Context(), id, platform.UITreeOptions{
		VisibleOnly: visibleOnly,
		TextFilter:  textFilter,
	})
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Failed to get UI tree: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, UITreeResponse{
		Tree:       tree,
		TokenCount: tree.TokenCount,
	})
}

// ElementsResponse is returned by GET /api/v1/devices/{id}/elements.
type ElementsResponse struct {
	Elements []platform.UIElement `json:"elements"`
	Count    int                  `json:"count"`
}

func (s *Server) handleFindElements(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("query"))
	textMatch := strings.TrimSpace(r.URL.Query().Get("text"))
	resIDMatch := strings.TrimSpace(r.URL.Query().Get("resource_id"))

	if query == "" && textMatch == "" && resIDMatch == "" {
		writeBadRequest(w, "At least one of query, text, or resource_id parameter is required")
		return
	}

	tree, err := adapter.GetUITree(r.Context(), id, platform.UITreeOptions{
		VisibleOnly: true,
	})
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Failed to inspect UI tree: %v", err))
		return
	}

	var matches []platform.UIElement
	if tree != nil {
		for _, el := range tree.Elements {
			matched := false
			if query != "" {
				if strings.Contains(strings.ToLower(el.Text), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(el.ResourceID), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(el.ContentDesc), strings.ToLower(query)) {
					matched = true
				}
			} else {
				if textMatch != "" && strings.Contains(strings.ToLower(el.Text), strings.ToLower(textMatch)) {
					matched = true
				}
				if resIDMatch != "" && strings.Contains(strings.ToLower(el.ResourceID), strings.ToLower(resIDMatch)) {
					matched = true
				}
			}

			if matched {
				matches = append(matches, el)
			}
		}
	}

	writeJSON(w, http.StatusOK, ElementsResponse{
		Elements: matches,
		Count:    len(matches),
	})
}
