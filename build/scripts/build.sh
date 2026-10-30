#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────
# Công cụ ngành Kiểm sát — build helper
# Build single-binary Windows/Linux với Wails v2 + Go.
# ─────────────────────────────────────────────────────────

set -e

export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
export GOPATH="$HOME/go"
export CGO_ENABLED=1

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TARGET="${1:-host}"
OUTPUT_DIR="$ROOT/build/bin"
mkdir -p "$OUTPUT_DIR"

# Wails v2 yêu cầu build tags "desktop,production" để bật webview binding.
# Trên Linux: link với libwebkit2gtk-4.1 (+gtk-3, gio, javascriptcoregtk).
# Trên Windows: dùng WebView2 (pre-installed trên Win10/11).

case "$TARGET" in
  linux|Linux)
    echo "==> Building Linux x86_64 binary"
    echo "    (requires libwebkit2gtk-4.1-dev libgtk-3-dev libsoup-3.0-dev)"
    CGO_ENABLED=1 CC=gcc GOOS=linux GOARCH=amd64 \
      go build -tags "desktop,production" \
        -ldflags "-s -w -X main.version=0.1.0" \
        -o "$OUTPUT_DIR/kiem_sat_toolkit_linux" .
    ls -lh "$OUTPUT_DIR/kiem_sat_toolkit_linux"
    echo ""
    echo "Run: $OUTPUT_DIR/kiem_sat_toolkit_linux"
    echo "    (libwebkit2gtk-4.1 phải có sẵn system-wide để chạy)"
    ;;
  windows|Windows|win)
    echo "==> Building Windows x86_64 binary (cross-compile)"
    echo "    (requires x86_64-w64-mingw32-gcc)"
    if ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then
      echo "ERROR: mingw-w64 not installed. Install with:"
      echo "  sudo apt install gcc-mingw-w64-x86-64"
      exit 1
    fi
    CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
      go build -tags "desktop,production" \
        -ldflags "-s -w -H windowsgui -X main.version=0.1.0" \
        -o "$OUTPUT_DIR/kiem_sat_toolkit.exe" .
    ls -lh "$OUTPUT_DIR/kiem_sat_toolkit.exe"
    echo ""
    echo "Run: $OUTPUT_DIR/kiem_sat_toolkit.exe"
    echo "    (uses WebView2 runtime — pre-installed on Windows 10/11)"
    ;;
  wails-linux)
    echo "==> Build via Wails CLI for Linux"
    wails build -platform linux/amd64 -clean
    ;;
  wails-windows)
    echo "==> Build via Wails CLI for Windows"
    wails build -platform windows/amd64 -clean
    ;;
  *)
    echo "Usage: $0 [linux|windows|wails-linux|wails-windows]"
    echo ""
    echo "Default: build for current host (with -tags desktop,production)"
    go build -tags "desktop,production" -ldflags "-s -w" \
      -o "$OUTPUT_DIR/kiem_sat_toolkit" .
    ls -lh "$OUTPUT_DIR/kiem_sat_toolkit"
    ;;
esac

echo ""
echo "Done."
