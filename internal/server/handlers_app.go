package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AppPackageRequest payload for launch, terminate, uninstall, and clear-data.
type AppPackageRequest struct {
	PackageName string `json:"package_name"`
}

// AppInstallRequest payload for install.
type AppInstallRequest struct {
	AppPath string `json:"app_path"`
}

// OpenURLRequest payload for open-url.
type OpenURLRequest struct {
	URL string `json:"url"`
}

func (s *Server) handleLaunchApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req AppPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}
	if strings.TrimSpace(req.PackageName) == "" {
		writeBadRequest(w, "package_name is required")
		return
	}

	res, err := adapter.LaunchApp(r.Context(), id, req.PackageName)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Launch app failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Launched application %s", req.PackageName)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

func (s *Server) handleTerminateApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req AppPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}
	if strings.TrimSpace(req.PackageName) == "" {
		writeBadRequest(w, "package_name is required")
		return
	}

	res, err := adapter.TerminateApp(r.Context(), id, req.PackageName)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Terminate app failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Terminated application %s", req.PackageName)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

func (s *Server) handleInstallApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req AppInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}
	if strings.TrimSpace(req.AppPath) == "" {
		writeBadRequest(w, "app_path is required")
		return
	}

	res, err := adapter.InstallApp(r.Context(), id, req.AppPath)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Install app failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Installed application from %s", req.AppPath)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

func (s *Server) handleUninstallApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req AppPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}
	if strings.TrimSpace(req.PackageName) == "" {
		writeBadRequest(w, "package_name is required")
		return
	}

	res, err := adapter.UninstallApp(r.Context(), id, req.PackageName)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Uninstall app failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Uninstalled application %s", req.PackageName)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

func (s *Server) handleClearAppData(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req AppPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}
	if strings.TrimSpace(req.PackageName) == "" {
		writeBadRequest(w, "package_name is required")
		return
	}

	res, err := adapter.ClearAppData(r.Context(), id, req.PackageName)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Clear app data failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Cleared data for application %s", req.PackageName)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}

func (s *Server) handleOpenURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adapter := s.registry.FindAdapter(r.Context(), id)
	if adapter == nil {
		writeNotFound(w, fmt.Sprintf("Device %q not found", id))
		return
	}

	var req OpenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, fmt.Sprintf("Invalid request JSON: %v", err))
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		writeBadRequest(w, "url is required")
		return
	}

	res, err := adapter.OpenURL(r.Context(), id, req.URL)
	if err != nil {
		writeInternalError(w, fmt.Sprintf("Open URL failed: %v", err))
		return
	}

	msg := fmt.Sprintf("Opened URL %s", req.URL)
	writeJSON(w, http.StatusOK, ActionResponse{
		Success:  res.Success,
		Message:  msg,
		Duration: res.Duration.String(),
	})
}
