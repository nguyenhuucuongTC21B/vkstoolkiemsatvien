package main

import (
        "embed"
        "log"
        "os"
        "path/filepath"

        "kiem_sat_toolkit/internal/db"
        "kiem_sat_toolkit/internal/modules"

        "github.com/wailsapp/wails/v2"
        "github.com/wailsapp/wails/v2/pkg/options"
        "github.com/wailsapp/wails/v2/pkg/options/assetserver"
        "github.com/wailsapp/wails/v2/pkg/options/linux"
        "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed assets/logo.png
var windowIcon []byte

// Lưu ý: icon file .exe Windows (icon.ico) được `wails build` tự động
// nhúng vào binary khi compile — không cần embed thủ công trong code.
// File `build/windows/icon.ico` phải tồn tại trước khi chạy `wails build`.

func main() {
        // User data dir: portable, next to the binary if possible, else ~/.kiem_sat_toolkit
        dataDir := resolveDataDir()
        if err := os.MkdirAll(dataDir, 0755); err != nil {
                log.Fatalf("cannot create data dir %s: %v", dataDir, err)
        }

        // Open the JSON-backed store
        dbMgr, err := db.NewManager(dataDir)
        if err != nil {
                log.Fatalf("db init: %v", err)
        }
        defer dbMgr.Close()

        // Build all modules
        registry := modules.NewRegistry(dbMgr, dataDir)
        app := NewApp(registry, dataDir)

        err = wails.Run(&options.App{
                Title:     "Công cụ ngành Kiểm sát",
                Width:     1440,
                Height:    900,
                MinWidth:  1024,
                MinHeight: 700,
                AssetServer: &assetserver.Options{
                        Assets: assets,
                },
                BackgroundColour: &options.RGBA{R: 244, G: 244, B: 245, A: 1},
                OnStartup:        app.OnStartup,
                OnDomReady:       app.OnDomReady,
                OnShutdown:       app.OnShutdown,
                Bind: []interface{}{
                        app,
                        registry.Calc,
                        registry.Legal,
                        registry.OCR,
                        registry.PII,
                        registry.Spell,
                        registry.File,
                        registry.Evidence,
                        registry.Case,
                        registry.Docs,
                },
                EnableDefaultContextMenu: true,
                // Linux: icon cửa sổ khi minimize + tên chương trình cho window manager
                Linux: &linux.Options{
                        Icon:        windowIcon,
                        ProgramName: "Công cụ ngành Kiểm sát",
                },
                // Windows: .exe icon sẽ tự lấy từ build/windows/icon.ico khi build bằng `wails build`
                Windows: &windows.Options{
                        WebviewIsTransparent: false,
                        WindowIsTranslucent:  false,
                },
        })

        if err != nil {
                log.Fatalf("wails run: %v", err)
        }
}

// resolveDataDir picks a writable data dir.
// Priority: ./data next to binary (portable), then $XDG_DATA_HOME, then ~/.kiem_sat_toolkit
func resolveDataDir() string {
        exe, err := os.Executable()
        if err == nil {
                portable := filepath.Join(filepath.Dir(exe), "data")
                if isWritable(portable) {
                        return portable
                }
        }
        if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
                return filepath.Join(xdg, "kiem_sat_toolkit")
        }
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".kiem_sat_toolkit")
}

func isWritable(dir string) bool {
        if err := os.MkdirAll(dir, 0755); err != nil {
                return false
        }
        tmp := filepath.Join(dir, ".wtest")
        f, err := os.Create(tmp)
        if err != nil {
                return false
        }
        f.Close()
        os.Remove(tmp)
        return true
}
