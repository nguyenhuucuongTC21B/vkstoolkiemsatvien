// Package audit provides a non-reversible audit logger shared by all
// module services. Keeping it in its own package avoids the import
// cycle that would otherwise form between `internal/modules` (registry)
// and `internal/modules/<name>` (each service).
package audit

import (
	"encoding/json"
	"fmt"
	"time"

	"kiem_sat_toolkit/internal/store"
)

// Logger writes audit entries to app store table "audit_log".
type Logger struct {
	tbl *store.Table
}

// New constructs a Logger bound to the given store.
func New(s *store.Store) *Logger {
	return &Logger{tbl: s.Table("audit_log")}
}

// Log records one audit entry. detail may be any JSON-marshalable value.
func (l *Logger) Log(module, action, actor string, detail interface{}, sensitivity string) {
	if l == nil || l.tbl == nil {
		return
	}
	detailJSON, _ := json.Marshal(detail)
	if sensitivity == "" {
		sensitivity = "normal"
	}
	row := store.Row{
		"ts":          time.Now().Format("2006-01-02 15:04:05"),
		"actor":       actor,
		"module":      module,
		"action":      action,
		"detail":      string(detailJSON),
		"sensitivity": sensitivity,
	}
	_ = l.tbl.Insert(row)
}

// Logf is a convenience wrapper for simple text details.
func (l *Logger) Logf(module, action, actor, format string, args ...interface{}) {
	l.Log(module, action, actor, map[string]string{"msg": fmt.Sprintf(format, args...)}, "normal")
}
