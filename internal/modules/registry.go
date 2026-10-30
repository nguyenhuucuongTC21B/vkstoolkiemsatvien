// Package modules aggregates all Wails-bound services used by the
// frontend. Each field is a separate binding so the UI can call
// `window.go.modules.Calc.AnPhiHS(...)` etc. directly.
package modules

import (
        "kiem_sat_toolkit/internal/audit"
        "kiem_sat_toolkit/internal/db"
        "kiem_sat_toolkit/internal/modules/calc"
        "kiem_sat_toolkit/internal/modules/casefile"
        "kiem_sat_toolkit/internal/modules/docs"
        "kiem_sat_toolkit/internal/modules/evidence"
        "kiem_sat_toolkit/internal/modules/file"
        "kiem_sat_toolkit/internal/modules/legal"
        "kiem_sat_toolkit/internal/modules/ocr"
        "kiem_sat_toolkit/internal/modules/pii"
        "kiem_sat_toolkit/internal/modules/spell"
)

// Registry holds all module instances. Wails binds each public field.
type Registry struct {
        Calc     *calc.Service
        Legal    *legal.Service
        OCR      *ocr.Service
        PII      *pii.Service
        Spell    *spell.Service
        File     *file.Service
        Evidence *evidence.Service
        Case     *casefile.Service
        Docs     *docs.Service
}

// NewRegistry wires up every module with its dependencies.
func NewRegistry(dbMgr *db.Manager, dataDir string) *Registry {
        auditLog := audit.New(dbMgr.App)

        return &Registry{
                Calc:     calc.New(dbMgr, auditLog),
                Legal:    legal.New(dbMgr, auditLog),
                OCR:      ocr.New(dbMgr, auditLog, dataDir),
                PII:      pii.New(auditLog),
                Spell:    spell.New(dbMgr, auditLog),
                File:     file.New(auditLog),
                Evidence: evidence.New(dbMgr, auditLog),
                Case:     casefile.New(dbMgr, auditLog),
                Docs:     docs.New(dbMgr, auditLog),
        }
}

// ModuleNames returns the public module identifiers — surfaced to
// the UI for the dashboard / sidebar.
func (r *Registry) ModuleNames() []string {
        return []string{
                "calc", "legal", "ocr", "pii", "spell",
                "file", "evidence", "case", "docs",
        }
}
