package main

import (
	"context"
	"log"
	"path/filepath"

	"kiem_sat_toolkit/internal/modules"
)

// App is the root binding exposed to the frontend.
// It provides metadata + a navigation/info API. Module-specific logic
// is bound separately (registry.Calc, registry.Legal, ...).
type App struct {
	ctx      context.Context
	registry *modules.Registry
	dataDir  string
}

func NewApp(reg *modules.Registry, dataDir string) *App {
	return &App{registry: reg, dataDir: dataDir}
}

// OnStartup is called when the Wails window starts.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("[INFO] data dir: %s", a.dataDir)
}

// OnDomReady is called when the frontend DOM is ready.
func (a *App) OnDomReady(ctx context.Context) {
	// no-op
}

// OnShutdown is called when the window is closing.
func (a *App) OnShutdown(ctx context.Context) {
	log.Printf("[INFO] shutdown")
}

// ────────────── Info API (exposed to frontend) ──────────────

// AppInfo describes the running application.
type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	DataDir     string `json:"dataDir"`
	Platform    string `json:"platform"`
	BuildDate   string `json:"buildDate"`
	OfflineOnly bool   `json:"offlineOnly"`
	Modules     []string `json:"modules"`
}

// GetAppInfo returns metadata about the running app.
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:        "Công cụ ngành Kiểm sát",
		Version:     "0.1.0-mvp",
		DataDir:     a.dataDir,
		Platform:    "linux",
		BuildDate:   "2026-10",
		OfflineOnly: true,
		Modules:     a.registry.ModuleNames(),
	}
}

// OpenDataDir returns the absolute path of the user data directory
// (where case files, exports, OCR output are stored).
func (a *App) OpenDataDir() string {
	return a.dataDir
}

// ExportDir returns the subdirectory used for exported files.
func (a *App) ExportDir() string {
	return filepath.Join(a.dataDir, "exports")
}

// CaseDir returns the subdirectory used for case files.
func (a *App) CaseDir() string {
	return filepath.Join(a.dataDir, "cases")
}
